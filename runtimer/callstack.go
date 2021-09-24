package runtimer

import (
	"fmt"
	"runtime"
)

type CallStack struct {
	Func string
	Line int
	Err  string
}

func (this *CallStack) ToString() string {
	if len(this.Err) < 1 {
		return fmt.Sprintf("{%v:%v}", this.Func, this.Line)
	}
	return fmt.Sprintf("{%v:%v|%v}", this.Func, this.Line, this.Err)
}

func GetCallStack(skip int, err error) *CallStack {
	// skip: 0 => self
	// skip: 1 => caller
	// skip: 2 => caller's caller
	pc, _, line, _ := runtime.Caller(skip)
	p := runtime.FuncForPC(pc)

	errStr := ""
	if nil != err {
		errStr = err.Error()
	}

	return &CallStack{p.Name(), line, errStr}
}

// MakeError
// @lastErr: should not be nil
// @newErr: could be nil
func MakeError(lastErr error, newErr error, skip int) error {
	callstack := GetCallStack(skip, newErr)
	return fmt.Errorf("%v|%v", callstack.ToString(), lastErr.Error())
}

func AppendError(lastErr error, newErr error) error {
	return MakeError(lastErr, newErr, 3)
}

func NewError(err error) error {
	return MakeError(err, nil, 3)
}
