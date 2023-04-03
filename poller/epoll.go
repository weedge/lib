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

var (
	epollFD int
)

func createPoller() (err error) {
	epollFD, err = syscall.EpollCreate1(0)
	if err != nil {
		return
	}
	return
}

func addReadEventFD(fd int) (err error) {
	err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: EpollRead,
		Fd:     int32(fd),
	})

	return
}

func delEventFD(fd int) error {
	err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}

	return nil
}

// todo: io_uring submit and wait_cqe, 
func getEvents() ([]event, error) {
	epollEvents := make([]syscall.EpollEvent, 100)
	n, err := syscall.EpollWait(epollFD, epollEvents, -1)
	if err != nil {
		return nil, err
	}

	events := make([]event, 0, len(epollEvents))
	for i := 0; i < n; i++ {
		event := event{
			FD: epollEvents[i].Fd, 
		}
		if epollEvents[i].Events == EpollClose {
			event.Type = EventClose
		} else {
			event.Type = EventIn
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
