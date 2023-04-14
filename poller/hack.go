//go:build linux
// +build linux

package poller

import (
	"syscall"
	_ "unsafe"
)

//go:linkname anyToSockaddr syscall.anyToSockaddr
func anyToSockaddr(rsa *syscall.RawSockaddrAny) (syscall.Sockaddr, error)
