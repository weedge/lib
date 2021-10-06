//@author wuyong
//@date   2020/8/21
//@desc

package chain

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type InitRoomActionCtx struct {
	LessonId int64
}

func (m *InitRoomActionCtx) ValidateData() (ok bool) {
	time.Sleep(3 * time.Second)
	fmt.Println("validate data")

	return true
}
func (m *InitRoomActionCtx) GetCurRoomInfos() (err error) {
	fmt.Println("getCurRoomInfos ok")

	return
}

func TestRunMethod(t *testing.T) {
	item := &InitRoomActionCtx{}
	handlerList := &Handlers{}
	vf := reflect.ValueOf(item)
	vft := vf.Type()
	mNum := vf.NumMethod()
	for i := 0; i < mNum; i++ {
		mName := vft.Method(i).Name
		reflectMethodValue := vf.Method(i)
		if !IsNoInParamAndOnlyErrorOutParamFunc(reflectMethodValue.Type()) {
			continue
		}

		if funcHandle, ok := vf.Method(i).Interface().(func() error); ok {
			handlerList.AddHandler(NewHandler(false, mName, funcHandle))
		}
	}
	err := handlerList.RunHandler()
	t.Logf("linkProcess cost %s", handlerList.FormatCost())
	if err != nil {
		t.Errorf("err[%s]", err.Error())
		return
	}
	t.Logf("ok")
}

// func () (err error) filter
func IsNoInParamAndOnlyErrorOutParamFunc(reflectType reflect.Type) (ok bool) {
	return reflectType.NumIn() == 0 && reflectType.NumOut() == 1 && reflectType.Out(0).String() == "error"
}

type GetRoomInfoActionCtx struct {
	LessonId   int64
	LiveRoomId int64
}

func (m *GetRoomInfoActionCtx) ValidateData() (ok bool) {
	time.Sleep(3 * time.Second)
	fmt.Println("validate data")

	return true
}
func (m *GetRoomInfoActionCtx) GetCurRoomInfos1() (err error) {
	fmt.Println("getCurRoomInfos1 ok")
	time.Sleep(1 * time.Second)

	return
}
func (m *GetRoomInfoActionCtx) GetCurRoomInfos2() (err error) {
	fmt.Println("getCurRoomInfos2 ok")
	time.Sleep(2 * time.Second)

	return
}
func (m *GetRoomInfoActionCtx) GetCurRoomInfos3() (err error) {
	fmt.Println("getCurRoomInfos3 ok")
	time.Sleep(3 * time.Second)

	return
}

func TestConcurrencyRunMethod(t *testing.T) {
	item := &GetRoomInfoActionCtx{}
	handlerList := &Handlers{}
	vf := reflect.ValueOf(item)
	vft := vf.Type()
	mNum := vf.NumMethod()
	for i := 0; i < mNum; i++ {
		mName := vft.Method(i).Name
		reflectMethodValue := vf.Method(i)
		if !IsNoInParamAndOnlyErrorOutParamFunc(reflectMethodValue.Type()) {
			continue
		}

		if funcHandle, ok := vf.Method(i).Interface().(func() error); ok {
			handlerList.AddHandler(NewHandler(false, mName, funcHandle))
		}
	}

	//concurHandlerList := NewConcurrencyHandlers(false, *handlerList)
	concurHandlerList := NewConcurrencyHandlers(true, *handlerList)
	err := concurHandlerList.RunHandler()
	t.Logf("handlerProcess cost %s", concurHandlerList.FormatCost())
	if err != nil {
		t.Errorf("err[%s]", err.Error())
		return
	}
	t.Logf("ok")
}
