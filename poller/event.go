package poller

import (
	"fmt"

	"github.com/ii64/gouring"
)

type eventInfo struct {
	fd    int                 // file desc no
	etype EventType           // event type
	bid   uint16              // buff id in pool group
	gid   uint16              // buff group id
	cb    EventCallBack       // callback
	cqe   *gouring.IoUringCqe // iouring complete queue entry for reap event
}

type EventType uint16
type EventCallBack func(info *eventInfo) error

const (
	ETypeUnknow  EventType = iota
	ETypeIn                // event stream ready to read
	ETypeClose             // close connect
	ETypeTimeout           // connect timeout

	ETypeAccept     // accept event op completed
	ETypeRead       // read event op completed
	ETypeWrite      // write event op completed
	ETypeProvidBuff // provide buff ok
	ETypePollInRead // kenerl poll in event ready
)

func (e *eventInfo) String() string {
	res := fmt.Sprintf(
		"fd:%d etype:%d bid:%d gid:%d cb:%v cqe:%v",
		e.fd, e.etype, e.bid, e.gid, e.cb, e.cqe)

	return res
}
