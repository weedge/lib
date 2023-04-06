//go:build linux
// +build linux

package poller

import (
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/ii64/gouring"
)

const (
	reqFeatures = gouring.IORING_FEAT_SINGLE_MMAP | gouring.IORING_FEAT_FAST_POLL | gouring.IORING_FEAT_NODROP
)

type ioUring struct {
	submitNum int64            // submit num for io_uring_enter submited check
	ring      *gouring.IoUring // liburing ring obj
}

// newIoUring
// new io uring with params, check required features,
// register ring fd
func newIoUring(entries uint32, params *gouring.IoUringParams) (iouring *ioUring, err error) {
	ring, err := gouring.NewWithParams(entries, params)
	if err != nil {
		return
	}

	if params.Features&reqFeatures == 0 {
		err = ErrIOUringFeaturesUnAvailable
		ring.Close()
		return
	}

	/*
		Note:
		When the ring descriptor is registered, it is stored internally in the struct io_uring structure.
		For applications that share a ring between threads, for example having one thread do submits and another reap events, then this optimization cannot be used as each thread may have a different index for the registered ring fd.
	*/
	ret, err := ring.RegisterRingFD()
	if err != nil || ret < 0 {
		err = ErrIOUringRegisterFDFail
		return
	}

	return &ioUring{ring: ring, submitNum: 0}, nil
}

func (m *ioUring) CloseRing() {
	m.ring.Close()
}

// getEventInfo io_uring submit and wait cqe for reap event
func (m *ioUring) getEventInfo() (info *eventInfo, err error) {
	submited, err := m.ring.Submit()
	if err != nil {
		return
	}
	if atomic.LoadInt64(&m.submitNum) != int64(submited) {
		err = ErrIOUringSubmitedNoFull
		return
	}
	atomic.StoreInt64(&m.submitNum, 0)

	var cqe *gouring.IoUringCqe
	err = m.ring.WaitCqe(&cqe)
	if err != nil {
		err = ErrIOUringWaitCqeFail
		return
	}

	info = (*eventInfo)(cqe.UserData.GetUnsafe())
	info.cqe = cqe

	return
}

// getEventInfos
// @todo
// io_uring submit and wait mutli cqe for reap events
func (m *ioUring) getEventInfos(infos []*eventInfo, err error) {

	return
}

func (m *ioUring) addAcceptSqe(cb EventCallBack, lfd int,
	clientAddr *syscall.RawSockaddrAny, clientAddrLen uint32, flags uint8) {
	sqe := m.ring.GetSqe()
	gouring.PrepAccept(sqe, lfd, clientAddr, (*uintptr)(unsafe.Pointer(&clientAddrLen)), 0)
	sqe.Flags = flags

	eventInfo := &eventInfo{
		fd:    lfd,
		etype: ETypeAccept,
		cb:    cb,
	}

	sqe.UserData.SetUnsafe(unsafe.Pointer(eventInfo))

	atomic.AddInt64(&m.submitNum, 1)
}

func (m *ioUring) addRecvSqe(cb EventCallBack, cfd int, buff []byte, flags uint8) {
	sqe := m.ring.GetSqe()
	gouring.PrepRecv(sqe, cfd, &buff[0], len(buff), uint(flags))
	sqe.Flags = flags

	eventInfo := eventInfo{
		fd:    cfd,
		etype: ETypeRead,
	}

	sqe.UserData.SetUnsafe(unsafe.Pointer(&eventInfo))
	atomic.AddInt64(&m.submitNum, 1)
}

func (m *ioUring) addSendSqe(cb EventCallBack, cfd int, buff []byte, msgSize int, flags uint8) {
	sqe := m.ring.GetSqe()
	gouring.PrepSend(sqe, cfd, &buff[0], msgSize, uint(flags))
	sqe.Flags = flags

	eventInfo := eventInfo{
		fd:    cfd,
		etype: ETypeWrite,
	}

	sqe.UserData.SetUnsafe(unsafe.Pointer(&eventInfo))
	atomic.AddInt64(&m.submitNum, 1)
}

func (m *ioUring) cqeDone(cqe *gouring.IoUringCqe) {
	m.ring.SeenCqe(cqe)
}
