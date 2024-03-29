//go:build linux
// +build linux

package poller

import (
	"time"

	"github.com/weedge/lib/log"
	"golang.org/x/sys/unix"
)

const (
	// man epoll_ctl  see EPOLL_EVENTS detail
	//1 2 8 16 8192 2147483648
	//EpollReadEvents = unix.EPOLLIN | unix.EPOLLPRI | unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP | unix.EPOLLET
	EpollReadEvents = unix.EPOLLIN | unix.EPOLLET
	EpollPeerClose  = unix.EPOLLIN | unix.EPOLLRDHUP
	EpollRead       = unix.EPOLLIN
	EpollErr        = unix.EPOLLIN | unix.EPOLLERR
	EpollReadErr    = unix.EPOLLIN | unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP
)

func createPoller() (pollFD int, err error) {
	pollFD, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return
	}
	return
}

func addReadEventFD(pollFD, fd int) (err error) {
	err = unix.EpollCtl(pollFD, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: EpollReadEvents,
		Fd:     int32(fd),
	})

	return
}

func delEventFD(pollFD, fd int) error {
	err := unix.EpollCtl(pollFD, unix.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}

	return nil
}

func getEvents(pollFD int) ([]eventInfo, error) {
	epollEvents := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(pollFD, epollEvents, -1)
	if err != nil {
		return nil, err
	}

	events := make([]eventInfo, 0, len(epollEvents))
	for i := 0; i < n; i++ {
		event := eventInfo{
			fd: int(epollEvents[i].Fd),
		}
		if epollEvents[i].Events == EpollRead { // likely
			event.etype = ETypeIn
			events = append(events, event)
		} else if epollEvents[i].Events == EpollPeerClose { // likely
			event.etype = ETypeClose
			events = append(events, event)
		} else if epollEvents[i].Events == EpollErr { // unlikely
			log.Errorf("epoll wait err event %v", epollEvents[i])
			event.etype = ETypeClose
			events = append(events, event)
		} else if epollEvents[i].Events == EpollReadErr { // unlikely
			log.Errorf("epoll wait read err event %v", epollEvents[i])
			event.etype = ETypeClose
			events = append(events, event)
		} else { // unlikely
			log.Errorf("epoll wait other event %v", epollEvents[i])
		}
	}
	//sort.Slice(events, func(i, j int) bool { return events[i].etype < events[j].etype })

	return events, nil
}

func setSockKeepAliveOptions(fd int, d time.Duration) (err error) {
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1)
	if err != nil {
		return
	}

	secs := int(roundDurationUp(d, time.Second))
	err = unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, secs)
	if err != nil {
		return
	}

	err = unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, secs)
	if err != nil {
		return
	}

	return
}
