package main

import (
	"time"
)

type Alarm struct {
	Hostname string
	control  chan int
	cancel   chan int
	manager  chan string
}

type Timer struct {
	Hostname string
	time     int
	res      chan int
}

type TimerManager struct {
	Update     chan Timer
	Deadhosts  chan string
	Cancelhost chan string
	Cancel     chan int
}

func managerproc(man *TimerManager) {
	managerchan := make(chan string)
	alarms := make(map[string]*Alarm)
	for {
		select {
		case updatemsg := <-man.Update:
			alarm, exists := alarms[updatemsg.Hostname]
			if exists {
				alarm.control <- updatemsg.time
				updatemsg.res <- 0
			} else {
				newalarm := &Alarm{
					Hostname: updatemsg.Hostname,
					control:  make(chan int),
					cancel:   make(chan int),
					manager:  managerchan,
				}
				alarms[updatemsg.Hostname] = newalarm
				go timerproc(newalarm)
				newalarm.control <- updatemsg.time
				updatemsg.res <- 0
			}
		case cancelhost := <-man.Cancelhost:
			alarm, exists := alarms[cancelhost]
			if exists {
				alarm.cancel <- 1
			}
		case deadhost := <-managerchan:
			deadalarm, exists := alarms[deadhost]
			if exists {
				delete(alarms, deadalarm.Hostname)
				man.Deadhosts <- deadalarm.Hostname
			}
		case <-man.Cancel:
			for _, value := range alarms {
				value.cancel <- 1
			}
			return
		}
	}
}

func timerproc(alarm *Alarm) {
	var timer *time.Timer = nil
	control := alarm.control
	cancel := alarm.cancel
	timer = time.NewTimer(time.Duration(<-control) * time.Second)
	for {
		select {
		case newinterval := <-control:
			// we're going to try the 1.23 timer updates
			//if timer != nil {
			//	timer.Stop()
			//}
			timer = time.NewTimer(time.Duration(newinterval) * time.Second)
			// defer timer.Stop()
		case <-cancel:
			timer.Stop()
			return
		case <-timer.C:
			alarm.manager <- alarm.Hostname
			return
		}
	}
}

func NewTimerManager() *TimerManager {
	newman := TimerManager{
		Update:     make(chan Timer),
		Deadhosts:  make(chan string),
		Cancelhost: make(chan string),
		Cancel:     make(chan int),
	}
	go managerproc(&newman)
	return &newman
}

func (tm *TimerManager) RegisterHost(host string, time int) int {
	nt := Timer{
		Hostname: host,
		time:     time,
		res:      make(chan int),
	}
	tm.Update <- nt
	return <-nt.res
}
