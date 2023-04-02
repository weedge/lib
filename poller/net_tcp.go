package poller

import (
	"fmt"
	"syscall"
	"time"

	"github.com/weedge/lib/log"
)

var (
	listenFD int
)

func listen(address string, backlog int) error {
	var err error
	listenFD, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Error(err)
		return err
	}
	err = syscall.SetsockoptInt(listenFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		log.Error(err)
		return err
	}

	addr, port, err := GetIPPort(address)
	if err != nil {
		return err
	}
	err = syscall.Bind(listenFD, &syscall.SockaddrInet4{
		Port: port,
		Addr: addr,
	})
	if err != nil {
		log.Error(err)
		return err
	}
	err = syscall.Listen(listenFD, backlog)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("listen addr", addr, "port", port)

	err = createPoller()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func accept(d time.Duration) (nfd int, addr string, err error) {
	nfd, sa, err := syscall.Accept(listenFD)
	if err != nil {
		return
	}

	// 设置为非阻塞状态
	err = syscall.SetNonblock(nfd, true)
	if err != nil {
		return
	}

	err = syscall.SetsockoptInt(nfd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
	if err != nil {
		return
	}

	if d == 0 {
		d = defaultTCPKeepAlive
	}
	err = setSockKeepAliveOptions(nfd, d)
	if err != nil {
		return
	}

	err = addRead(nfd)
	if err != nil {
		return
	}
	addr = getAddr(sa)

	return
}

func addRead(fd int) (err error) {
	err = addReadEventFD(fd)
	if err != nil {
		return
	}

	return
}

func closeFD(fd int) (err error) {
	err = delEventFD(fd)
	if err != nil {
		return
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
