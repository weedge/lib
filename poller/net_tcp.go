package poller

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

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
	err = delEventFD(pollerFD, fd)
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

func anyToSockaddr(rsa *syscall.RawSockaddrAny) (syscall.Sockaddr, error) {
	switch rsa.Addr.Family {
	case syscall.AF_NETLINK:
		pp := (*syscall.RawSockaddrNetlink)(unsafe.Pointer(rsa))
		sa := new(syscall.SockaddrNetlink)
		sa.Family = pp.Family
		sa.Pad = pp.Pad
		sa.Pid = pp.Pid
		sa.Groups = pp.Groups
		return sa, nil

	case syscall.AF_PACKET:
		pp := (*syscall.RawSockaddrLinklayer)(unsafe.Pointer(rsa))
		sa := new(syscall.SockaddrLinklayer)
		sa.Protocol = pp.Protocol
		sa.Ifindex = int(pp.Ifindex)
		sa.Hatype = pp.Hatype
		sa.Pkttype = pp.Pkttype
		sa.Halen = pp.Halen
		sa.Addr = pp.Addr
		return sa, nil

	case syscall.AF_UNIX:
		pp := (*syscall.RawSockaddrUnix)(unsafe.Pointer(rsa))
		sa := new(syscall.SockaddrUnix)
		if pp.Path[0] == 0 {
			// "Abstract" Unix domain socket.
			// Rewrite leading NUL as @ for textual display.
			// (This is the standard convention.)
			// Not friendly to overwrite in place,
			// but the callers below don't care.
			pp.Path[0] = '@'
		}

		// Assume path ends at NUL.
		// This is not technically the Linux semantics for
		// abstract Unix domain sockets--they are supposed
		// to be uninterpreted fixed-size binary blobs--but
		// everyone uses this convention.
		n := 0
		for n < len(pp.Path) && pp.Path[n] != 0 {
			n++
		}
		bytes := (*[len(pp.Path)]byte)(unsafe.Pointer(&pp.Path[0]))[0:n]
		sa.Name = string(bytes)
		return sa, nil

	case syscall.AF_INET:
		pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(rsa))
		sa := new(syscall.SockaddrInet4)
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		sa.Port = int(p[0])<<8 + int(p[1])
		sa.Addr = pp.Addr
		return sa, nil

	case syscall.AF_INET6:
		pp := (*syscall.RawSockaddrInet6)(unsafe.Pointer(rsa))
		sa := new(syscall.SockaddrInet6)
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		sa.Port = int(p[0])<<8 + int(p[1])
		sa.ZoneId = pp.Scope_id
		sa.Addr = pp.Addr
		return sa, nil
	}
	return nil, syscall.EAFNOSUPPORT
}
