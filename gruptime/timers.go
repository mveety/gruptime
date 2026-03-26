package main

import (
	"errors"
	"time"
)

const (
	TimerOpBad       = -1
	TimerOpRemaining = iota
)

type AlarmMessage struct {
	resp chan time.Time
}

type Alarm struct {
	Hostname string
	control  chan time.Duration
	cancel   chan int
	manager  chan string
	endtime  time.Time
	endtimes chan AlarmMessage
}

type Timer struct {
	Hostname string
	time     time.Duration
	res      chan int
}

type TimerResponse struct {
	op     int
	time   time.Time
	status error
}

type TimerRequest struct {
	op    int
	timer string
	resp  chan TimerResponse
}

type TimerManager struct {
	Update     chan Timer
	Deadhosts  chan string
	Cancelhost chan string
	Cancel     chan int
	Requests   chan TimerRequest
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
					control:  make(chan time.Duration),
					cancel:   make(chan int),
					manager:  managerchan,
					endtimes: make(chan AlarmMessage),
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
		case req := <-man.Requests:
			switch req.op {
			default:
				req.resp <- TimerResponse{
					op:     TimerOpBad,
					status: errors.New("bad request"),
				}
			case TimerOpRemaining:
				alarm, exists := alarms[req.timer]
				if !exists {
					req.resp <- TimerResponse{
						op:     TimerOpRemaining,
						status: errors.New("timer missing"),
					}
				} else {
					endtime := alarm.getendtime()
					req.resp <- TimerResponse{
						op:     TimerOpRemaining,
						time:   endtime,
						status: nil,
					}
				}
			}
		}
	}
}

func (alarm *Alarm) getendtime() time.Time {
	msg := AlarmMessage{resp: make(chan time.Time)}
	alarm.endtimes <- msg
	endtime := <-msg.resp
	return endtime
}

func timerproc(alarm *Alarm) {
	var timer *time.Timer = nil
	control := alarm.control
	cancel := alarm.cancel
	firstduration := <-control
	alarm.endtime = time.Now().Add(firstduration)
	timer = time.NewTimer(firstduration)
	for {
		select {
		case newinterval := <-control:
			timer.Stop()
			alarm.endtime = time.Now().Add(newinterval)
			timer.Reset(newinterval)
		case <-cancel:
			timer.Stop()
			return
		case <-timer.C:
			alarm.manager <- alarm.Hostname
			return
		case msg := <-alarm.endtimes:
			currentendtime := alarm.endtime
			msg.resp <- currentendtime
		}
	}
}

func NewTimerManager() *TimerManager {
	newman := TimerManager{
		Update:     make(chan Timer),
		Deadhosts:  make(chan string),
		Cancelhost: make(chan string),
		Cancel:     make(chan int),
		Requests:   make(chan TimerRequest),
	}
	go managerproc(&newman)
	return &newman
}

func (tm *TimerManager) RegisterHost(host string, time time.Duration) int {
	nt := Timer{
		Hostname: host,
		time:     time,
		res:      make(chan int),
	}
	tm.Update <- nt
	return <-nt.res
}

func (tm *TimerManager) EndTime(host string) (time.Time, error) {
	req := TimerRequest{
		op:    TimerOpRemaining,
		timer: host,
		resp:  make(chan TimerResponse),
	}
	tm.Requests <- req
	resp := <-req.resp
	if resp.op != TimerOpRemaining {
		panic("invalid operation in response")
	}
	return resp.time, resp.status
}
