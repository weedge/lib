//go:build linux
// +build linux

package poller

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/ii64/gouring"
	"github.com/weedge/lib/log"
)

const (
	reqFeatures = gouring.IORING_FEAT_SINGLE_MMAP | gouring.IORING_FEAT_FAST_POLL | gouring.IORING_FEAT_NODROP
)

type ioUring struct {
	//submitNum    int64            // submit num for io_uring_enter submited check
	spins             int64                           // spins count for submit and wait timeout
	ring              *gouring.IoUring                // liburing ring obj
	submitSignal      chan struct{}                   // submit signal
	mapUserDataEvent  map[gouring.UserData]*eventInfo // user data from cqe to event info
	userDataEventLock sync.RWMutex                    // rwlock for mapUserDataEvent
}

// newIoUring
// new io uring with params, check required features,
// register ring fd
func newIoUring(entries uint32, params *gouring.IoUringParams) (iouring *ioUring, err error) {
	ring, err := gouring.NewWithParams(entries, params)
	if err != nil {
		return
	}

	if params != nil && params.Features&reqFeatures == 0 {
		err = ErrIOUringFeaturesUnAvailable
		ring.Close()
		return
	}

	/*
			Note:
			When the ring descriptor is registered, it is stored internally in the struct io_uring structure.
			For applications that share a ring between threads, for example having one thread do submits and another reap events, then this optimization cannot be used as each thread may have a different index for the registered ring fd.

		ret, err := ring.RegisterRingFD()
		if err != nil || ret < 0 {
			log.Errorf("ring.RegisterRingFD err %s", err.Error())
			err = ErrIOUringRegisterFDFail
			return
		}
	*/

	log.Infof("newIoUring ok")

	return &ioUring{
		ring:             ring,
		submitSignal:     make(chan struct{}),
		mapUserDataEvent: make(map[gouring.UserData]*eventInfo),
	}, nil
}

func (m *ioUring) CloseRing() {
	if m.ring != nil {
		m.ring.Close()
	}
}

// getEventInfo io_uring submit and wait cqe for reap event
// notice: gc
func (m *ioUring) getEventInfo() (info *eventInfo, err error) {
	if atomic.AddInt64(&m.spins, 1) <= 20 {
		return
	}
	atomic.StoreInt64(&m.spins, 0)

	var cqeData *gouring.IoUringCqe
	// submit wait at least 1 cqe and wait 1 us timeout, todo: use sync call instead of async callback
	err = m.ring.SubmitAndWaitTimeOut(&cqeData, 1, 1, nil)
	if err != nil {
		if err == syscall.ETIME || err == syscall.EINTR {
			err = nil
		}
		return
	}
	if cqeData == nil {
		return
	}

	cqe := *cqeData
	if cqe.UserData.GetUnsafe() == nil {
		// Own timeout doesn't have user data
		errStr := fmt.Sprintf("no user data, cqe:%+v", cqe)
		err = errors.New(errStr)
		return
	}

	m.userDataEventLock.Lock()
	info = m.mapUserDataEvent[cqe.UserData]
	//info = (*eventInfo)(cqe.UserData.GetUnsafe())
	if info != nil && (info.cb == nil || info.etype == ETypeUnknow) {
		m.userDataEventLock.Unlock()
		err = errors.New("error event infoPtr")
		return
	}
	//https://github.com/golang/go/issues/20135
	delete(m.mapUserDataEvent, cqe.UserData)
	m.userDataEventLock.Unlock()

	info.cqe = cqe

	log.Infof("get event info: %s", info)

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
	println("addAcceptSqe")
	sqe := m.ring.GetSqe()
	gouring.PrepAccept(sqe, lfd, clientAddr, (*uintptr)(unsafe.Pointer(&clientAddrLen)), 0)
	sqe.Flags = flags

	eventInfo := &eventInfo{
		fd:    lfd,
		etype: ETypeAccept,
		cb:    cb,
	}

	sqe.UserData.SetUnsafe(unsafe.Pointer(eventInfo))
	m.userDataEventLock.Lock()
	m.mapUserDataEvent[sqe.UserData] = eventInfo
	m.userDataEventLock.Unlock()
}

func (m *ioUring) addRecvSqe(cb EventCallBack, cfd int, buff []byte, size int, flags uint8) {
	println("addRecvSqe")
	var buf *byte
	if len(buff) > 0 {
		buf = &buff[0]
	}
	sqe := m.ring.GetSqe()
	gouring.PrepRecv(sqe, cfd, buf, size, uint(flags))
	sqe.Flags = flags

	eventInfo := &eventInfo{
		fd:    cfd,
		etype: ETypeRead,
		cb:    cb,
	}

	sqe.UserData.SetUnsafe(unsafe.Pointer(eventInfo))
	m.userDataEventLock.Lock()
	m.mapUserDataEvent[sqe.UserData] = eventInfo
	m.userDataEventLock.Unlock()
}

func (m *ioUring) addSendSqe(cb EventCallBack, cfd int, buff []byte, msgSize int, flags uint8) {
	println("addSendSqe")
	var buf *byte
	if len(buff) > 0 {
		buf = &buff[0]
	}
	sqe := m.ring.GetSqe()
	gouring.PrepSend(sqe, cfd, buf, msgSize, uint(flags))
	sqe.Flags = flags

	eventInfo := &eventInfo{
		fd:    cfd,
		etype: ETypeWrite,
		cb:    cb,
	}

	sqe.UserData.SetUnsafe(unsafe.Pointer(eventInfo))
	m.userDataEventLock.Lock()
	m.mapUserDataEvent[sqe.UserData] = eventInfo
	m.userDataEventLock.Unlock()
}

func (m *ioUring) cqeDone(cqe gouring.IoUringCqe) {
	m.ring.SeenCqe(&cqe)
}
