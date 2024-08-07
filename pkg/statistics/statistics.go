package statistics

import (
	"Open_IM/pkg/common/log"
	"time"
)

type Statistics struct {
	AllCount   *uint64
	ModuleName string
	PrintArgs  string
	SleepTime  uint64
}

func (s *Statistics) output() {
	var intervalCount int64
	t := time.NewTicker(time.Duration(s.SleepTime) * time.Second)
	defer t.Stop()
	var sum uint64
	var timeIntervalNum uint64
	for {
		sum = *s.AllCount
		select {
		case <-t.C:
		}
		if *s.AllCount-sum <= 0 {
			intervalCount = 0
		} else {
			intervalCount = int64(*s.AllCount - sum)
		}
		timeIntervalNum++
		//system stat  									   msg_gateway  60 second add user conn 0     total:   4             intervalNum 5022               avg 0
		log.NewWarn("", " system stat ", s.ModuleName, s.PrintArgs, intervalCount, "total:", *s.AllCount, "intervalNum", timeIntervalNum, "avg", (*s.AllCount)/(timeIntervalNum)/s.SleepTime)
	}
}

func NewStatistics(allCount *uint64, moduleName, printArgs string, sleepTime int) *Statistics {
	p := &Statistics{AllCount: allCount, ModuleName: moduleName, SleepTime: uint64(sleepTime), PrintArgs: printArgs}
	go p.output()
	return p
}
