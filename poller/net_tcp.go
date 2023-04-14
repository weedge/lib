package poller

import (
	"fmt"
	"syscall"
	"time"

	"github.com/weedge/lib/log"
)

func listen(address string, backlog int) (listenFD int, err error) {
	listenFD, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Error(err)
		return
	}

	err = syscall.SetsockoptInt(listenFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		log.Error(err)
		return
	}

	addr, port, err := GetIPPort(address)
	if err != nil {
		return
	}

	err = syscall.Bind(listenFD, &syscall.SockaddrInet4{
		Port: port,
		Addr: addr,
	})
	if err != nil {
		log.Error(err)
		return
	}

	err = syscall.Listen(listenFD, backlog)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("server listen port %d fd %d", port, listenFD)

	return
}

func accept(listenFD int, d time.Duration) (nfd int, sa syscall.Sockaddr, err error) {
	// block accept
	nfd, sa, err = syscall.Accept(listenFD)
	if err != nil {
		return
	}

	err = setConnectOption(nfd, d)
	if err != nil {
		return
	}

	return
}

func setConnectOption(nfd int, d time.Duration) (err error) {
	// set connect fd non bolock
	err = syscall.SetNonblock(nfd, true)
	if err != nil {
		return
	}

	// set nodelay for wide bound network
	err = syscall.SetsockoptInt(nfd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
	if err != nil {
		return
	}

	// set tcp server connect keep alive
	if d == 0 {
		d = defaultTCPKeepAlive
	}
	err = setSockKeepAliveOptions(nfd, d)
	if err != nil {
		return
	}

	return
}

func addReadEvent(pollerFD, fd int) (err error) {
	err = addReadEventFD(pollerFD, fd)
	if err != nil {
		return
	}

	return
}

func closeFD(pollerFD, fd int) (err error) {
	if pollerFD > 0 {
		err = delEventFD(pollerFD, fd)
		if err != nil {
			return
		}
	}

	err = syscall.Close(fd)
	if err != nil {
		return
	}

	return
}

func getAddr(sa syscall.Sockaddr) string {
	addr, ok := sa.(*syscall.SockaddrInet4)
	if !ok {
		return ""
	}

	return fmt.Sprintf("%d.%d.%d.%d:%d", addr.Addr[0], addr.Addr[1], addr.Addr[2], addr.Addr[3], addr.Port)
}

func closeFDRead(fd int) error {
	_, _, e := syscall.Syscall(syscall.SHUT_RD, uintptr(fd), 0, 0)
	if e != 0 {
		return e
	}
	return nil
}