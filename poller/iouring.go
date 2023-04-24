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
	"golang.org/x/sys/unix"
)

const (
	reqFeatures = gouring.IORING_FEAT_SINGLE_MMAP | gouring.IORING_FEAT_FAST_POLL | gouring.IORING_FEAT_NODROP
)

type ioUring struct {
	spins             int64                           // spins count for submit and wait timeout
	ring              *gouring.IoUring                // liburing ring obj
	eventfd           int                             // register iouring eventfd
	mapUserDataEvent  map[gouring.UserData]*eventInfo // user data from cqe to event info
	userDataEventLock sync.RWMutex                    // rwlock for mapUserDataEvent
	subLock           sync.Mutex
	cqeSignCh         chan struct{}
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

	iouring = &ioUring{
		ring:             ring,
		mapUserDataEvent: make(map[gouring.UserData]*eventInfo),
		cqeSignCh:        make(chan struct{}, 1),
	}

	return
}

func (m *ioUring) CloseRing() {
	if m.ring != nil {
		m.ring.Close()
	}
}

func (m *ioUring) RegisterEventFd() (err error) {
	eventfd, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		return
	}
	m.eventfd = eventfd

	err = m.ring.RegisterEventFd(m.eventfd)
	if err != nil {
		return
	}

	return
}

// getEventInfo io_uring submit and wait cqe for reap event
// notice: gc
func (m *ioUring) getEventInfov1() (info *eventInfo, err error) {
	if atomic.AddInt64(&m.spins, 1) <= 20 {
		return
	}
	atomic.StoreInt64(&m.spins, 0)

	var cqeData *gouring.IoUringCqe
	// submit wait at least 1 cqe and wait 1 us timeout, todo: use sync call instead of async callback
	err = m.ring.SubmitAndWaitTimeOut(&cqeData, 1, 1, nil)
	if err != nil {
		if errors.Is(err, syscall.ETIME) || errors.Is(err, syscall.EINTR) || errors.Is(err, syscall.EAGAIN) {
			err = nil
		}
		return
	}
	if cqeData == nil {
		return
	}

	if cqeData.UserData.GetUnsafe() == nil {
		// Own timeout doesn't have user data
		errStr := fmt.Sprintf("no user data, cqe:%+v", cqeData)
		err = errors.New(errStr)
		return
	}
	cqe := *cqeData

	m.userDataEventLock.Lock()
	info, ok := m.mapUserDataEvent[cqe.UserData]
	if !ok {
		errStr := fmt.Sprintf("cqe %+v userData %d get event info: %s empty", cqe, cqe.UserData, info)
		m.userDataEventLock.Unlock()
		//panic(errStr)
		log.Error(errStr)
		// commit cqe is seen
		m.cqeDone(cqe)
		return
	}
	//info = (*eventInfo)(cqe.UserData.GetUnsafe())
	if info != nil && (info.cb == nil || info.etype == ETypeUnknow) {
		m.userDataEventLock.Unlock()
		err = errors.New("error event infoPtr")
		// commit cqe is seen
		m.cqeDone(cqe)
		return
	}
	//https://github.com/golang/go/issues/20135
	delete(m.mapUserDataEvent, cqe.UserData)
	log.Infof("userData %d get event info: %s", cqe.UserData, info)
	info.cqe = cqe
	m.userDataEventLock.Unlock()

	return
}

// getEventInfos
// @todo
// io_uring submit and wait mutli cqe for reap events
func (m *ioUring) getEventInfos(infos []*eventInfo, err error) {
	return
}

func (m *ioUring) getEventInfo() (info *eventInfo, err error) {
	for {
		select {
		case <-m.cqeSignCh:
			var cqe *gouring.IoUringCqe
			err = m.ring.PeekCqe(&cqe)
			if err != nil {
				continue
			}
			if cqe == nil {
				log.Warnf("cqe is nil")
				continue
			}

			m.userDataEventLock.Lock()
			info, ok := m.mapUserDataEvent[cqe.UserData]
			if !ok {
				errStr := fmt.Sprintf("cqe %+v userData %d get event info: %s empty", cqe, cqe.UserData, info)
				m.userDataEventLock.Unlock()
				//panic(errStr)
				log.Error(errStr)
				// commit cqe is seen
				m.cqeDone(*cqe)
				continue
			}
			//info = (*eventInfo)(cqe.UserData.GetUnsafe())
			if info != nil && (info.cb == nil || info.etype == ETypeUnknow) {
				m.userDataEventLock.Unlock()
				log.Error("error event infoPtr")
				// commit cqe is seen
				m.cqeDone(*cqe)
				continue
			}
			//https://github.com/golang/go/issues/20135
			delete(m.mapUserDataEvent, cqe.UserData)
			log.Debugf("userData %d get event info: %s", cqe.UserData, info)
			info.cqe = *cqe
			m.userDataEventLock.Unlock()

			return info, nil
		}
	}
}

func (m *ioUring) addAcceptSqe(cb EventCallBack, lfd int,
	clientAddr *syscall.RawSockaddrAny, clientAddrLen uint32, flags uint8) {
	m.subLock.Lock()
	defer m.subLock.Unlock()
	sqe := m.ring.GetSqe()
	gouring.PrepAccept(sqe, lfd, clientAddr, (*uintptr)(unsafe.Pointer(&clientAddrLen)), 0)
	sqe.Flags = flags

	eventInfo := &eventInfo{
		fd:    lfd,
		etype: ETypeAccept,
		cb:    cb,
	}

	sqe.UserData = gouring.UserData(uintptr(unsafe.Pointer(eventInfo)))
	m.userDataEventLock.Lock()
	m.mapUserDataEvent[sqe.UserData] = eventInfo
	m.userDataEventLock.Unlock()
	log.Debugf("addAcceptSqe userData %d eventInfo:%s", sqe.UserData, eventInfo)
	m.ring.Submit()
}

func (m *ioUring) addRecvSqe(cb EventCallBack, cfd int, buff []byte, size int, flags uint8) {
	m.subLock.Lock()
	defer m.subLock.Unlock()
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

	//sqe.UserData.SetUnsafe(unsafe.Pointer(eventInfo))
	sqe.UserData = gouring.UserData(uintptr(unsafe.Pointer(eventInfo)))
	m.userDataEventLock.Lock()
	m.mapUserDataEvent[sqe.UserData] = eventInfo
	m.userDataEventLock.Unlock()
	log.Debugf("addRecvSqe userData %d eventInfo:%s", sqe.UserData, eventInfo)
	m.ring.Submit()
}

func (m *ioUring) addSendSqe(cb EventCallBack, cfd int, buff []byte, msgSize int, flags uint8) {
	m.subLock.Lock()
	defer m.subLock.Unlock()
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

	//sqe.UserData.SetUnsafe(unsafe.Pointer(eventInfo))
	sqe.UserData = gouring.UserData(uintptr(unsafe.Pointer(eventInfo)))
	m.userDataEventLock.Lock()
	m.mapUserDataEvent[sqe.UserData] = eventInfo
	m.userDataEventLock.Unlock()
	log.Debugf("addSendSqe userData %d eventInfo:%s", sqe.UserData, eventInfo)
	m.ring.Submit()
}

func (m *ioUring) cqeDone(cqe gouring.IoUringCqe) {
	m.ring.SeenCqe(&cqe)
}
