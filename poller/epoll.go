//go:build linux
// +build linux

package poller

import (
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	EpollRead  = unix.EPOLLIN | unix.EPOLLPRI | unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLET | unix.EPOLLRDHUP
	EpollClose = uint32(unix.EPOLLIN | unix.EPOLLRDHUP)
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
		Events: EpollRead,
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
		if epollEvents[i].Events == EpollClose {
			event.etype = ETypeClose
		} else {
			event.etype = ETypeIn
		}
		events = append(events, event)
	}

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
