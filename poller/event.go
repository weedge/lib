package poller

import (
	"fmt"
	"time"

	"github.com/ii64/gouring"
)

type eventInfo struct {
	fd      int                // file desc no
	etype   EventType          // event type
	bid     uint16             // buff id in pool group
	gid     uint16             // buff group id
	cb      EventCallBack      // callback
	cqe     gouring.IoUringCqe // iouring complete queue entry for reap event
	timeOut time.Duration
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

var noOpsEventCb = func(info *eventInfo) error { return nil }

func (e *eventInfo) String() string {
	res := ""
	if e != nil {
		res = fmt.Sprintf("fd:%d etype:%d bid:%d gid:%d cb:%v cqe:%v",
			e.fd, e.etype, e.bid, e.gid, e.cb, e.cqe)
	}

	return res
}

/*
// no reflect.DeepCopy issue detail: https://github.com/golang/go/issues/51520
// so use the 3rd  go-clone for function
func (e *eventInfo) Clone() (newOne *eventInfo) {
	goClone.SetCustomFunc(reflect.TypeOf(eventInfo{}), func(allocator *goClone.Allocator, old, new reflect.Value) {
		oldField := old.FieldByName("cb")
		newField := new.FieldByName("cb")
		//f := unsafe.Pointer(uintptr(unsafe.Pointer(e)) + unsafe.Offsetof(e.cb))

		newField.SetPointer(oldField.UnsafePointer())
	})

	newOne = goClone.Clone(e).(*eventInfo)
	return
}
*/
