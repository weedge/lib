package tapper

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

type logUnit struct {
	key   string
	value string
}

type LogUnits struct {
	cmd          string
	begTime      int64
	units        []logUnit
	timeCostUnit []logUnit
	lock         sync.Mutex
}

func (p *LogUnits) SetCmd(cmd string) {
	p.cmd = cmd
}

func (p *LogUnits) SetBeginTime() {
	p.begTime = time.Now().UnixNano()
}

// add log unit like bd odp addNotice
func (p *LogUnits) AddLogUnit(k string, v string) {
	p.units = append(p.units, logUnit{key: k, value: v})
}

func (p *LogUnits) AddTimeCost(k string, tNs int64) {
	p.timeCostUnit = append(p.timeCostUnit, logUnit{key: k, value: strconv.FormatInt(tNs/1000000, 10)})
}

// multi go routine safe add
func (p *LogUnits) AddLogUnitThreadSafe(k string, v string) {
	p.lock.Lock()
	p.units = append(p.units, logUnit{key: k, value: v})
	p.lock.Unlock()
}

// multi go routine 版本
func (p *LogUnits) AddTimeCostThreadSafe(k string, tNs int64) {
	p.lock.Lock()
	p.timeCostUnit = append(p.timeCostUnit, logUnit{key: k, value: strconv.FormatInt(tNs/1000000, 10)})
	p.lock.Unlock()
}

func (p *LogUnits) String() string {
	endTime := time.Now().UnixNano()
	//var str string
	var buf strings.Builder
	buf.WriteString(p.cmd)
	buf.WriteString("||cost=")
	buf.WriteString(strconv.FormatInt((endTime-p.begTime)/int64(time.Millisecond), 10))
	for idx := range p.units {
		buf.WriteString(" ")
		buf.WriteString(p.units[idx].key)
		buf.WriteString("[")
		buf.WriteString(p.units[idx].value)
		buf.WriteString("]")
	}
	for idx := range p.timeCostUnit {
		buf.WriteString("||")
		buf.WriteString(p.timeCostUnit[idx].key)
		buf.WriteString("=")
		buf.WriteString(p.timeCostUnit[idx].value)
	}
	return buf.String()
}

func (p *LogUnits) SerializeTimeCost() string {
	var buf strings.Builder
	buf.WriteString("[")
	for idx := range p.timeCostUnit {
		buf.WriteString("{\"")
		buf.WriteString(p.timeCostUnit[idx].key)
		buf.WriteString("\":")
		buf.WriteString(p.timeCostUnit[idx].value)
		buf.WriteString("}")
		if idx < len(p.timeCostUnit)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func (p *LogUnits) MergeLogUnit(o *LogUnits, prefix string) {
	p.lock.Lock()
	for i := range o.units {
		unit := o.units[i]
		p.units = append(p.units, logUnit{key: prefix + unit.key, value: unit.value})
	}

	for i := range o.timeCostUnit {
		timeCostUnit := o.timeCostUnit[i]
		p.timeCostUnit = append(p.timeCostUnit, logUnit{key: prefix + timeCostUnit.key, value: timeCostUnit.value})
	}
	p.lock.Unlock()
}
