// +build darwin dragonfly freebsd netbsd openbsd

package poller

import (
	"golang.org/x/sys/unix"
)

var (
	kqueueFD int
)

const (
	// EV_ADD adds the event to the kqueue. Re-adding an existing event will modify
	// the parameters of the original event, and not result in a duplicate
	// entry. Adding an event automatically enables it, unless overridden by
	// the EV_DISABLE flag.
	EV_ADD = unix.EV_ADD

	// EV_ENABLE permits kevent() to return the event if it is triggered.
	EV_ENABLE = unix.EV_ENABLE

	// EV_DISABLE disables the event so kevent() will not return it. The filter itself is
	// not disabled.
	EV_DISABLE = unix.EV_DISABLE

	// EV_DISPATCH disables the event source immediately after delivery of an event. See
	// EV_DISABLE above.
	EV_DISPATCH = unix.EV_DISPATCH

	// EV_DELETE removes the event from the kqueue. Events which are attached to file
	// descriptors are automatically deleted on the last close of the
	// descriptor.
	EV_DELETE = unix.EV_DELETE

	// EV_RECEIPT is useful for making bulk changes to a kqueue without draining
	// any pending events. When passed as input, it forces EV_ERROR to always
	// be returned. When a filter is successfully added the data field will be
	// zero.
	EV_RECEIPT = unix.EV_RECEIPT

	// EV_ONESHOT causes the event to return only the first occurrence of the
	// filter being triggered. After the user retrieves the event from the
	// kqueue, it is deleted.
	EV_ONESHOT = unix.EV_ONESHOT

	// EV_CLEAR makes event state be reset after the event is retrieved by the
	// user. This is useful for filters which report state transitions instead
	// of the current state. Note that some filters may automatically set this
	// flag internally.
	EV_CLEAR = unix.EV_CLEAR

	// EV_EOF may be set by the filters to indicate filter-specific EOF
	// condition.
	EV_EOF = unix.EV_EOF

	// EV_ERROR is set to indiacate an error occured with the identtifier.
	EV_ERROR = unix.EV_ERROR
)

const (
	// EVFILT_READ takes a descriptor as the identifier, and returns whenever
	// there is data available to read. The behavior of the filter is slightly
	// different depending on the descriptor type.
	EVFILT_READ = unix.EVFILT_READ

	// EVFILT_WRITE takes a descriptor as the identifier, and returns whenever
	// it is possible to write to the descriptor. For sockets, pipes and fifos,
	// data will contain the amount of space remaining in the write buffer. The
	// filter will set EV_EOF when the reader disconnects, and for the fifo
	// case, this may be cleared by use of EV_CLEAR. Note that this filter is
	// not supported for vnodes or BPF devices. For sockets, the low water mark
	// and socket error handling is identical to the EVFILT_READ case.
	EVFILT_WRITE = unix.EVFILT_WRITE

	// EVFILT_AIO the sigevent portion of the AIO request is filled in, with
	// sigev_notify_kqueue containing the descriptor of the kqueue that the
	// event should be attached to, sigev_notify_kevent_flags containing the
	// kevent flags which should be EV_ONESHOT, EV_CLEAR or EV_DISPATCH,
	// sigev_value containing the udata value, and sigev_notify set to
	// SIGEV_KEVENT. When the aio_*() system call is made, the event will be
	// registered with the specified kqueue, and the ident argument set to the
	// struct aiocb returned by the aio_*() system call. The filter returns
	// under the same conditions as aio_error().
	EVFILT_AIO = unix.EVFILT_AIO

	// EVFILT_VNODE takes a file descriptor as the identifier and the events to
	// watch for in fflags, and returns when one or more of the requested
	// events occurs on the descriptor.
	EVFILT_VNODE = unix.EVFILT_VNODE

	// EVFILT_PROC takes the process ID to monitor as the identifier and the
	// events to watch for in fflags, and returns when the process performs one
	// or more of the requested events. If a process can normally see another
	// process, it can attach an event to it.
	EVFILT_PROC = unix.EVFILT_PROC

	// EVFILT_SIGNAL takes the signal number to monitor as the identifier and
	// returns when the given signal is delivered to the process. This coexists
	// with the signal() and sigaction() facilities, and has a lower
	// precedence. The filter will record all attempts to deliver a signal to
	// a process, even if the signal has been marked as SIG_IGN, except for the
	// SIGCHLD signal, which, if ignored, won't be recorded by the filter.
	// Event notification happens after normal signal delivery processing. data
	// returns the number of times the signal has occurred since the last call
	// to kevent(). This filter automatically sets the EV_CLEAR flag
	// internally.
	EVFILT_SIGNAL = unix.EVFILT_SIGNAL

	// EVFILT_TIMER establishes an arbitrary timer identified by ident. When
	// adding a timer, data specifies the timeout period. The timer will be
	// periodic unless EV_ONESHOT is specified. On return, data contains the
	// number of times the timeout has expired since the last call to kevent().
	// This filter automatically sets the EV_CLEAR flag internally. There is a
	// system wide limit on the number of timers which is controlled by the
	// kern.kq_calloutmax sysctl.
	EVFILT_TIMER = unix.EVFILT_TIMER

	// EVFILT_USER establishes a user event identified by ident which is not
	// associated with any kernel mechanism but is trig- gered by user level
	// code.
	EVFILT_USER = unix.EVFILT_USER

	// Custom filter value signaling that kqueue instance get closed.
	_EVFILT_CLOSED = -0x7f
)

const (
	EVRead  = EV_ADD | EV_ENABLE
	EVClose = EV_EOF | EV_DELETE
)

func createPoller() (err error) {
	kqueueFD, err = unix.Kqueue()
	if err != nil {
		return
	}

	return
}

func addReadEventFD(fd int) (err error) {
	changeEvent := unix.Kevent_t{
		Ident:  uint64(fd),
		Filter: int16(EVFILT_READ),
		Flags:  EVRead,
	}
	_, err = unix.Kevent(kqueueFD, []unix.Kevent_t{changeEvent}, nil, nil)
	if err != nil {
		return
	}

	return
}

func delEventFD(fd int) (err error) {
	_, err = unix.Kevent(kqueueFD,
		[]unix.Kevent_t{{Ident: uint64(fd), Flags: EV_DELETE}},
		nil, nil)

	return
}

func getEvents() ([]event, error) {
	kEvents := make([]unix.Kevent_t, 100)
	n, err := unix.Kevent(kqueueFD, nil, kEvents, nil)
	if err != nil {
		return nil, err
	}

	events := make([]event, 0, len(kEvents))
	for i := 0; i < n; i++ {
		event := event{
			FD: int32(kEvents[i].Ident),
		}
		if kEvents[i].Flags == EV_EOF {
			event.Type = EventClose
		} else {
			event.Type = EventIn
		}
		events = append(events, event)
	}

	return events, nil
}
