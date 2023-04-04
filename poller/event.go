package poller

type eventInfo struct {
	fd    int       // file desc no
	etype EventType // event type
	bid   uint16    // buff id in pool group
	gid   uint16    // buff group id
}

type EventType uint16

const (
	ETypeUnknow  EventType = iota
	ETypeIn                // event stream ready to read
	ETypeClose             // close connect
	ETypeTimeout           // connect timeout

	ETypeAccept     // accept event op completed
	ETypeRead       // read event op completed
	ETypeWrite      // write event op completed
	ETypeProvidBuff // provide buff ok
)
