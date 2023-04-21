//go:build linux
// +build linux

package poller

import (
	"syscall"
	"time"

	"github.com/weedge/lib/log"
	"golang.org/x/sys/unix"
)

const (
	// man epoll_ctl  see EPOLL_EVENTS detail
	EpollReadEvents = unix.EPOLLIN | unix.EPOLLPRI | unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLET | unix.EPOLLRDHUP
	EpollPeerClose  = unix.EPOLLIN | unix.EPOLLRDHUP
	EpollErr        = unix.EPOLLIN | unix.EPOLLERR
	EpollRead       = unix.EPOLLIN | unix.EPOLLET
)

func createPoller() (pollFD int, err error) {
	pollFD, err = syscall.EpollCreate1(0)
	if err != nil {
		return
	}
	return
}

func addReadEventFD(pollFD, fd int) (err error) {
	err = syscall.EpollCtl(pollFD, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: EpollReadEvents,
		Fd:     int32(fd),
	})

	return
}

func delEventFD(pollFD, fd int) error {
	err := syscall.EpollCtl(pollFD, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}

	return nil
}

func getEvents(pollFD int) ([]eventInfo, error) {
	epollEvents := make([]syscall.EpollEvent, 100)
	n, err := syscall.EpollWait(pollFD, epollEvents, -1)
	if err != nil {
		return nil, err
	}

	events := make([]eventInfo, 0, len(epollEvents))
	for i := 0; i < n; i++ {
		event := eventInfo{
			fd: int(epollEvents[i].Fd),
		}
		if epollEvents[i].Events == EpollPeerClose {
			event.etype = ETypeClose
		} else if epollEvents[i].Events == EpollErr {
			log.Errorf("epoll wait err %d event %v", EpollErr, epollEvents[i])
		} else if epollEvents[i].Events == EpollRead {
			event.etype = ETypeIn
		} else {
			//event.etype = ETypeIn
		}
		events = append(events, event)
	}
	//sort.Slice(events, func(i, j int) bool { return events[i].etype < events[j].etype })

	return events, nil
}

func setSockKeepAliveOptions(fd int, d time.Duration) (err error) {
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)
	if err != nil {
		return
	}

	secs := int(roundDurationUp(d, time.Second))
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs)
	if err != nil {
		return
	}

	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs)
	if err != nil {
		return
	}

	return
}
