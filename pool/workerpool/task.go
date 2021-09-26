package workerpool

import (
	"time"
)

type Task struct {
	Do          func(inParam interface{}, outParam interface{}) bool //返回值主要来知道是否执行完
	InParam     interface{}                                          //输入
	OutParam    interface{}                                          //输出
	ChIsTimeOut chan<- bool                                          //是否超时，true 超时,false 正常
	TimeOut     time.Duration                                        //任务超时时间
	OnTimeOut   func(inParam interface{}, outParam interface{})
}
