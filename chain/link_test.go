//@author wuyong
//@date   2020/7/7
//@desc

package chain

import (
	"fmt"
	"testing"
)

type PreloadBaseData struct {
	StrategyType int
	StrategyId   int64
	CourseId     int64
	LessonId     int64
	LecturerUid  int64
}

type PreloadRoomData struct {
	*PreloadBaseData
	LiveRoomId int64
}

type PreloadOrgData struct {
	*PreloadBaseData
	NodeId int64
}

type PreloadUserData struct {
	*PreloadRoomData
	*PreloadOrgData
	Uid int64
}

func (data *PreloadBaseData) ValidateData() (ok bool) {

	return true
}

func (data *PreloadBaseData) Process(input IParam) (err error, output IParam) {
	var ok bool
	if data, ok = input.(*PreloadBaseData); !ok {
		return fmt.Errorf("preload init data: %v fail~! input data: %v", data, input), input
	}

	if !data.ValidateData() {
		return fmt.Errorf("preload base data: %v validate fail~! input data: %v", data, input), input
	}

	output = &PreloadRoomData{
		PreloadBaseData: data,
		LiveRoomId:      100000,
	}

	fmt.Printf("preload base data: %v validate ok~! input data: %v output data: %v \n", data, input, output)

	return
}

func (data *PreloadRoomData) ValidateData() (ok bool) {

	return true
}

func (data *PreloadRoomData) Process(input IParam) (err error, output IParam) {
	var ok bool
	if data, ok = input.(*PreloadRoomData); !ok {
		return fmt.Errorf("preload init data: %v fail~! input data: %v", data, input), input
	}

	if !data.ValidateData() {
		return fmt.Errorf("room data: %v validate fail~! input data: %v", data, input), input
	}

	output = &PreloadOrgData{
		PreloadBaseData: data.PreloadBaseData,
		NodeId:          111111,
	}

	fmt.Printf("preload room data: %v validate ok~! input data: %v output data: %v \n", data, input, output)

	return
}

func (data *PreloadOrgData) ValidateData() (ok bool) {

	return true
}

func (data *PreloadOrgData) Process(input IParam) (err error, output IParam) {
	var ok bool
	if data, ok = input.(*PreloadOrgData); !ok {
		return fmt.Errorf("preload init data: %v fail~! input data: %v", data, input), input
	}

	if !data.ValidateData() {
		return fmt.Errorf("org data: %v validate fail~! input data: %v", data, input), input
	}

	output = &PreloadUserData{
		PreloadRoomData: &PreloadRoomData{
			PreloadBaseData: data.PreloadBaseData,
			LiveRoomId:      100000,
		},
		PreloadOrgData: &PreloadOrgData{
			PreloadBaseData: data.PreloadBaseData,
			NodeId:          111111,
		},
		Uid: 1213123123,
	}

	fmt.Printf("preload org data: %v validate ok~! input data: %v output data: %v \n", data, input, output)

	return
}

func (data *PreloadUserData) ValidateData() (ok bool) {

	return true
}

func (data *PreloadUserData) Process(input IParam) (err error, output IParam) {
	var ok bool
	if data, ok = input.(*PreloadUserData); !ok {
		return fmt.Errorf("preload init data: %v fail~! input data: %v", data, input), input
	}

	if !data.ValidateData() {
		return fmt.Errorf("user data: %v validate fail~! input data: %v", data, input), input
	}
	output = input

	fmt.Printf("preload user data: %v validate ok~! input data: %v output data: %v \n", data, input, output)

	return
}

func TestMain(m *testing.M) {
	m.Run()
}

func TestLinkHandle(t *testing.T) {
	baseData := &PreloadBaseData{StrategyType: 1, StrategyId: 2, CourseId: 1, LessonId: 2, LecturerUid: 3}
	linkHeadItem := NewLinkItem(false, "base", baseData)
	link := InitLink(linkHeadItem)
	t.Logf("link:%v linkItem:%v", link, linkHeadItem)

	roomData := &PreloadRoomData{baseData, 10}
	linkItem := NewLinkItem(true, "room", roomData)
	link.SetNextItem(linkItem)
	t.Logf("link:%v linkItem:%v", link, linkItem)

	orgData := &PreloadOrgData{baseData, 110}
	linkItem = NewLinkItem(false, "org", orgData)
	link.SetNextItem(linkItem)
	t.Logf("link:%v linkItem:%v", link, linkItem)

	userData := &PreloadUserData{roomData, orgData, 123}
	linkItem = NewLinkItem(true, "user", userData)
	link.SetNextItem(linkItem)
	t.Logf("link:%v linkItem:%v", link, linkItem)

	err := link.CheckSameName()
	if err != nil {
		t.Errorf("error: %v", err)
	}

	err = link.Handle(baseData)
	t.Logf("linkProcess cost %s", link.FormatCost())
	if err != nil {
		t.Errorf("error: %v", err)
	}

	return
}
