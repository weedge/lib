//@author wuyong
//@date   2020/7/7
//@desc   simple link processor

package chain

import (
	"fmt"
	"reflect"
	"time"

	"github.com/weedge/lib/log"
	//JsonIter "github.com/json-iterator/go"
)

type IParam interface {
	ValidateData() bool
}

type IProcessor interface {
	Process(input IParam) (err error, output IParam)
}

const (
	PROCESS_STATUS_UNDO = iota
	PROCESS_STATUS_DOING
	PROCESS_STATUS_OK
	PROCESS_STATUS_FAIL
)

// ctx->ctx->ctx
type LinkItem struct {
	IsPass    bool
	Id        int
	Name      string
	Cost      int64
	Processor IProcessor
	Status    int
	Next      *LinkItem
}

type Link struct {
	Head   *LinkItem
	Tail   *LinkItem
	Length int
}

func NewLinkItem(isPass bool, name string, processor IProcessor) (linkItem *LinkItem) {
	return &LinkItem{IsPass: isPass, Name: name, Processor: processor, Status: PROCESS_STATUS_UNDO}
}

func (item *LinkItem) SetNext(linkItem *LinkItem) {
	item.Next = linkItem
}

func InitLink(head *LinkItem) (link *Link) {
	if head == nil {
		return
	}

	return &Link{Head: head, Tail: head, Length: 1}
}

func (link *Link) SetNextItem(linkItem *LinkItem) {
	if linkItem == nil {
		return
	}
	linkItem.Id = link.Length
	link.Tail.SetNext(linkItem)
	link.Length += 1
	link.Tail = link.Tail.Next
}

func (l *Link) Handle(input IParam) (err error) {
	var output IParam
	var preTime int64
	p := l.Head
	for p != nil {
		log.Infof("id[%d]_%s_input_%s[%v]", p.Id, p.Name, reflect.TypeOf(input).String(), input)

		preTime = time.Now().UnixNano()
		p.Status = PROCESS_STATUS_DOING
		err, output = p.Processor.Process(input)
		p.Cost = int64((time.Now().UnixNano() - preTime) / 1000000)
		if err != nil {
			p.Status = PROCESS_STATUS_FAIL
			log.Errorf("%s process err", p.Name)
			if !p.IsPass {
				return
			}
			log.Infof("%s process pass", p.Name)
		} else {
			p.Status = PROCESS_STATUS_OK
			log.Infof("%s process ok", p.Name)
		}

		if p.Next == nil { //end print output
			log.Infof("%s_output[%v]", p.Name, output)
		}

		input = output
		p = p.Next
	} //end for

	return
}

func (l *Link) CheckSameName() (err error) {
	mapName := map[string]struct{}{}
	p := l.Head
	for p != nil {
		if _, ok := mapName[p.Name]; ok {
			return fmt.Errorf("have the same name: %s", p.Name)
		}
		p = p.Next
	}
	return
}

func (l *Link) GetAllProcessorCost() (mapCost map[string]int64) {
	mapCost = map[string]int64{}
	p := l.Head
	for p != nil {
		mapCost[p.Name] = p.Cost
		p = p.Next
	}

	return
}

func (l *Link) FormatCost() (str string) {
	totalCost := int64(0)
	p := l.Head
	for p != nil {
		str += fmt.Sprintf("%s_pass[%t]_status[%d]_cost[%d]->", p.Name, p.IsPass, p.Status, p.Cost)
		totalCost += p.Cost
		p = p.Next
	}
	str += fmt.Sprintf("totalCost:%d", totalCost)
	return
}

func (l *Link) HandleMap(input map[string]interface{}) (err error) {
	return
}

func (l *Link) HandleRpc(input []byte) (err error) {
	return
}
