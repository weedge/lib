package chain

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/weedge/lib/log"
)

type ICtx interface {
	ValidateData() (err error)
}
type HandleFunc func() error
type HandleCtxFunc func(ctx ICtx) error

type Handler struct {
	IsPass       bool
	Name         string
	Cost         int64
	Handle       HandleFunc
	CtxHandle    HandleCtxFunc
	ReflectValue reflect.Value
	Status       int
}

func NewHandler(isPass bool, name string, funcHandle HandleFunc) Handler {
	return Handler{IsPass: isPass, Name: name, Handle: funcHandle}
}

func NewCtxHandler(isPass bool, name string, funcHandle HandleCtxFunc) Handler {
	return Handler{IsPass: isPass, Name: name, CtxHandle: funcHandle}
}

func NewReflectHandler(isPass bool, name string, reflectValue reflect.Value) Handler {
	return Handler{IsPass: isPass, Name: name, ReflectValue: reflectValue}
}

type Handlers []Handler

func (s Handlers) Len() int           { return len(s) }
func (s Handlers) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Handlers) Less(i, j int) bool { return s[i].Cost < s[j].Cost }

type ConcurrencyHandlers struct {
	Concurrency bool
	Handlers    Handlers
}

func NewConcurrencyHandlers(concurrency bool, handlers Handlers) *ConcurrencyHandlers {
	return &ConcurrencyHandlers{Concurrency: concurrency, Handlers: handlers}
}

func (m *ConcurrencyHandlers) RunHandler() (err error) {
	if m.Concurrency {
		return m.Handlers.ConcurrencyRunHandler()
	}

	return m.Handlers.RunHandler()
}

func (m *ConcurrencyHandlers) RunCtxHandler(ctx ICtx) (err error) {
	if m.Concurrency {
		return m.Handlers.ConcurrencyRunCtxHandler(ctx)
	}

	return m.Handlers.RunCtxHandler(ctx)
}

func (m *ConcurrencyHandlers) FormatCost() (str string) {
	var buf strings.Builder

	if m.Concurrency {
		buf.WriteString("concurrency:")
		buf.WriteString(m.Handlers.FormatMaxCost())
	} else {
		buf.WriteString(m.Handlers.FormatCost())
	}

	return buf.String()
}

func (m *Handlers) AddHandler(handler ...Handler) {
	*m = append(*m, handler...)
}

func (m *Handlers) RunCtxHandler(ctx ICtx) (err error) {
	err = ctx.ValidateData()
	if err != nil {
		return
	}
	for index, item := range *m {
		preTime := time.Now().UnixNano()
		(*m)[index].Status = PROCESS_STATUS_DOING
		err = item.CtxHandle(ctx)
		(*m)[index].Cost = (time.Now().UnixNano() - preTime) / 1000000
		if err != nil {
			(*m)[index].Status = PROCESS_STATUS_FAIL
			log.Errorf("%s process ctx[%v] err[%s]", item.Name, ctx, err.Error())
			if !item.IsPass {
				err = fmt.Errorf("%s process ctx[%v] err[%s]", item.Name, ctx, err.Error())
				return
			}
			log.Infof("%s process ctx[%v] error[%s] pass", item.Name, ctx, err.Error())
		} else {
			(*m)[index].Status = PROCESS_STATUS_OK
			log.Infof("%s process ctx[%v] ok", item.Name, ctx)
		}
	}

	return
}

func (m *Handlers) RunHandler() (err error) {
	for index, item := range *m {
		preTime := time.Now().UnixNano()
		(*m)[index].Status = PROCESS_STATUS_DOING
		err = item.Handle()
		(*m)[index].Cost = (time.Now().UnixNano() - preTime) / 1000000
		if err != nil {
			(*m)[index].Status = PROCESS_STATUS_FAIL
			log.Errorf("%s process err[%s]", item.Name, err.Error())
			if !item.IsPass {
				err = fmt.Errorf("%s process err[%s]", item.Name, err.Error())
				return
			}
			log.Infof("%s process error[%s] pass", item.Name, err.Error())
		} else {
			(*m)[index].Status = PROCESS_STATUS_OK
			log.Infof("%s process ok", item.Name)
		}
	}

	return
}

func (m *Handlers) ConcurrencyRunCtxHandler(ctx ICtx) (err error) {
	err = ctx.ValidateData()
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	for index, item := range *m {
		wg.Add(1)
		go func(wg *sync.WaitGroup, index int, item Handler) {
			defer wg.Done()
			preTime := time.Now().UnixNano()
			(*m)[index].Status = PROCESS_STATUS_DOING
			localErr := item.CtxHandle(ctx)
			(*m)[index].Cost = int64((time.Now().UnixNano() - preTime) / 1000000)
			if localErr != nil {
				(*m)[index].Status = PROCESS_STATUS_FAIL
				log.Errorf("%s process ctx[%v] err[%s]", item.Name, ctx, localErr.Error())
				if !item.IsPass {
					err = fmt.Errorf("%s process ctx[%v] err[%s]", item.Name, ctx, err.Error())
					return
				}
				log.Infof("%s process ctx[%v] error[%s] pass", item.Name, ctx, localErr.Error())
			} else {
				(*m)[index].Status = PROCESS_STATUS_OK
				log.Infof("%s process ctx[%v] ok", item.Name, ctx)
			}
		}(&wg, index, item)
	}
	wg.Wait()

	return
}

func (m *Handlers) ConcurrencyRunHandler() (err error) {
	var wg sync.WaitGroup
	for index, item := range *m {
		wg.Add(1)
		go func(wg *sync.WaitGroup, index int, item Handler) {
			defer wg.Done()
			preTime := time.Now().UnixNano()
			(*m)[index].Status = PROCESS_STATUS_DOING
			err = item.Handle()
			(*m)[index].Cost = int64((time.Now().UnixNano() - preTime) / 1000000)
			if err != nil {
				(*m)[index].Status = PROCESS_STATUS_FAIL
				log.Errorf("%s process err[%s]", item.Name, err.Error())
				if !item.IsPass {
					err = fmt.Errorf("%s process err[%s]", item.Name, err.Error())
					return
				}
				log.Infof("%s process error[%s] pass", item.Name, err.Error())
			} else {
				(*m)[index].Status = PROCESS_STATUS_OK
				log.Infof("%s process ok", item.Name)
			}
		}(&wg, index, item)
	}
	wg.Wait()

	return
}

func (m *Handlers) RunReflectValueCall() (err error) {
	for index, item := range *m {
		preTime := time.Now().UnixNano()
		(*m)[index].Status = PROCESS_STATUS_DOING
		reflectValues := item.ReflectValue.Call([]reflect.Value{})
		(*m)[index].Cost = int64((time.Now().UnixNano() - preTime) / 1000000)
		if len(reflectValues) == 0 {
			(*m)[index].Status = PROCESS_STATUS_FAIL
			err = fmt.Errorf("refect value len is empty")
			return
		}
		var ok bool
		err, ok = reflectValues[0].Interface().(error)
		if ok {
			(*m)[index].Status = PROCESS_STATUS_FAIL
			log.Errorf("%s process err[%s]", item.Name, err.Error())
			if !item.IsPass {
				err = fmt.Errorf("%s process err[%s]", item.Name, err.Error())
				return
			}
			log.Infof("%s process error[%s] pass", item.Name, err.Error())
		} else {
			(*m)[index].Status = PROCESS_STATUS_OK
			log.Infof("%s process ok", item.Name)
		}
	}

	return
}

func (m *Handlers) FormatCost() (str string) {
	totalCost := int64(0)
	for _, item := range *m {
		str += fmt.Sprintf("%s_pass[%t]_status[%d]_cost[%d]->", item.Name, item.IsPass, item.Status, item.Cost)
		totalCost += item.Cost
	}
	str += fmt.Sprintf("totalCost:%d", totalCost)
	return
}

func (m *Handlers) FormatMaxCost() (str string) {
	sort.Sort(m)
	for _, item := range *m {
		str += fmt.Sprintf("%s_pass[%t]_status[%d]_cost[%d]||", item.Name, item.IsPass, item.Status, item.Cost)
	}
	str += fmt.Sprintf("totalCost:%d", (*m)[m.Len()-1].Cost)
	return
}
