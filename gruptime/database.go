package main

import (
	"errors"
	"github.com/mveety/gruptime/internal/uptime"
)

var (
	HostTimeout     = 480 // hosts timeout every 480 seconds
	ErrDbNotStarted = errors.New("db not started")
	ErrDbStarted    = errors.New("db already started")
	ErrNoHost       = errors.New("host not found")
)

const (
	OpAddHost = iota
	OpGetHost
	OpRemoveHost
	OpGetAllHosts
)

type DbResponse struct {
	err error
	one uptime.Uptime
	all []uptime.Uptime
}

type DbMessage struct {
	op       int
	resp     chan DbResponse
	hostname string
	data     uptime.Uptime
}

type Database struct {
	data     map[string]uptime.Uptime
	commchan chan DbMessage
	timers   *TimerManager
}

func initUptimedb() *Database {
	newdb := new(Database)
	newdb.data = make(map[string]uptime.Uptime)
	newdb.commchan = make(chan DbMessage)
	newdb.timers = NewTimerManager()
	go dbproc(newdb)
	return newdb
}

func (d *Database) handlemessage(msg DbMessage) {
	switch msg.op {
	case OpAddHost:
		d.data[msg.data.Hostname] = msg.data
		d.timers.RegisterHost(msg.data.Hostname, HostTimeout)
		msg.resp <- DbResponse{err: nil}
		return
	case OpGetHost:
		uptime, exists := d.data[msg.hostname]
		if exists {
			msg.resp <- DbResponse{err: nil, one: uptime}
		} else {
			msg.resp <- DbResponse{err: ErrNoHost}
		}
		return
	case OpRemoveHost:
		_, exists := d.data[msg.hostname]
		if exists {
			delete(d.data, msg.hostname)
			d.timers.Cancelhost <- msg.hostname
			msg.resp <- DbResponse{err: nil}
		} else {
			msg.resp <- DbResponse{err: ErrNoHost}
		}
		return
	case OpGetAllHosts:
		size := len(d.data)
		if size < 1 {
			msg.resp <- DbResponse{err: ErrNoHost}
			return
		}
		uptimes := make([]uptime.Uptime, size)
		i := 0
		for _, u := range d.data {
			uptimes[i] = u
			i++
		}
		msg.resp <- DbResponse{err: nil, all: uptimes}
		return
	}
}

func dbproc(db *Database) {
	for {
		select {
		case msg := <-db.commchan:
			db.handlemessage(msg)
		case deadhost := <-db.timers.Deadhosts:
			_, exists := db.data[deadhost]
			if exists {
				delete(db.data, deadhost)
			}
		}
	}
}

func (d *Database) AddHost(host string, ut uptime.Uptime) error {
	msg := DbMessage{
		op:       OpAddHost,
		resp:     make(chan DbResponse),
		hostname: host,
		data:     ut,
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.err
}

func (d *Database) GetHost(host string) (uptime.Uptime, error) {
	msg := DbMessage{
		op:       OpGetHost,
		resp:     make(chan DbResponse),
		hostname: host,
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.one, res.err
}

func (d *Database) RemoveHost(host string) error {
	msg := DbMessage{
		op:       OpRemoveHost,
		resp:     make(chan DbResponse),
		hostname: host,
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.err
}

func (d *Database) GetAllHosts() ([]uptime.Uptime, error) {
	msg := DbMessage{
		op:   OpGetAllHosts,
		resp: make(chan DbResponse),
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.all, res.err
}
