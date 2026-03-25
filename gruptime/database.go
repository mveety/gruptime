package main

import (
	"errors"
	"log"
	"maps"
	"strings"
	"time"

	"github.com/mveety/gruptime/internal/uptime"
)

var (
	ErrDBNotStarted = errors.New("db not started")
	ErrDBStarted    = errors.New("db already started")
	ErrNoHost       = errors.New("host not found")
)

const (
	OpAddHost = iota
	OpGetHost
	OpRemoveHost
	OpRemovePeer
	OpGetAllHosts
	OpGetAllPeers
)

type DBResponse struct {
	err   error
	one   uptime.Uptime
	all   []uptime.Uptime
	peers map[string]bool
}

type DBMessage struct {
	op       int
	resp     chan DBResponse
	hostname string
	data     uptime.Uptime
}

type Database struct {
	data       map[string]uptime.Uptime
	peers      map[string]bool
	commchan   chan DBMessage
	timers     *TimerManager
	peertimers *TimerManager
}

func initUptimedb() *Database {
	newdb := Database{
		data:       make(map[string]uptime.Uptime),
		peers:      make(map[string]bool),
		commchan:   make(chan DBMessage),
		timers:     NewTimerManager(),
		peertimers: NewTimerManager(),
	}
	go dbproc(&newdb)
	return &newdb
}

func (d *Database) handlemessage(msg DBMessage) {
	switch msg.op {
	case OpAddHost:
		newuptime := msg.data
		timeout := time.Duration(runningConfig.HostTimeout) * time.Second
		if msg.data.Lifetime > 0 {
			timeout = newuptime.Lifetime
		} else {
			newuptime.Lifetime = time.Duration(runningConfig.HostTimeout) * time.Second
		}
		endtime, err := d.timers.EndTime(newuptime.Hostname)
		if err != nil && newuptime.Lifetime < time.Until(endtime) {
			if runningConfig.Verbose {
				log.Printf("dropping shorter lived uptime for %s", newuptime.Hostname)
			}
			msg.resp <- DBResponse{err: nil}
			return
		}
		hostname := strings.Clone(newuptime.Hostname)
		d.data[hostname] = newuptime
		d.timers.RegisterHost(hostname, timeout)
		d.peers[hostname] = true
		d.peertimers.RegisterHost(hostname, time.Duration(runningConfig.PeerTimeout)*time.Second)
		msg.resp <- DBResponse{err: nil}
		return
	case OpGetHost:
		uptime, exists := d.data[msg.hostname]
		if exists {
			endtime, err := d.timers.EndTime(msg.hostname)
			if err != nil {
				panic(err)
			}
			uptime.Lifetime = time.Until(endtime)
			msg.resp <- DBResponse{err: nil, one: uptime}
		} else {
			msg.resp <- DBResponse{err: ErrNoHost}
		}
		return
	case OpRemoveHost:
		_, exists := d.data[msg.hostname]
		if exists {
			delete(d.data, msg.hostname)
			d.timers.Cancelhost <- msg.hostname
			_, exists = d.peers[msg.hostname]
			if exists {
				d.peers[msg.hostname] = false
			}
			msg.resp <- DBResponse{err: nil}
		} else {
			msg.resp <- DBResponse{err: ErrNoHost}
		}
		return
	case OpRemovePeer:
		_, exists := d.peers[msg.hostname]
		if exists {
			delete(d.peers, msg.hostname)
			d.peertimers.Cancelhost <- msg.hostname
			msg.resp <- DBResponse{err: nil}
		} else {
			msg.resp <- DBResponse{err: ErrNoHost}
		}
	case OpGetAllHosts:
		size := len(d.data)
		if size < 1 {
			msg.resp <- DBResponse{err: ErrNoHost}
			return
		}
		curdata := maps.Clone(d.data)
		uptimes := make([]uptime.Uptime, size)
		i := 0
		for _, u := range curdata {
			endtime, err := d.timers.EndTime(u.Hostname)
			if err != nil {
				panic(err)
			}
			u.Lifetime = time.Until(endtime)
			uptimes[i] = u
			i++
		}
		msg.resp <- DBResponse{err: nil, all: uptimes}
		return
	case OpGetAllPeers:
		size := len(d.peers)
		if size < 1 {
			msg.resp <- DBResponse{err: ErrNoHost}
		}
		msg.resp <- DBResponse{err: nil, peers: maps.Clone(d.peers)}
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
			_, exists = db.peers[deadhost]
			if exists {
				db.peers[deadhost] = false
			}
		case deadpeer := <-db.peertimers.Deadhosts:
			_, exists := db.peers[deadpeer]
			if exists {
				delete(db.data, deadpeer)
			}
		}
	}
}

func (d *Database) AddHost(host string, ut uptime.Uptime) error {
	msg := DBMessage{
		op:       OpAddHost,
		resp:     make(chan DBResponse),
		hostname: host,
		data:     ut,
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.err
}

func (d *Database) GetHost(host string) (uptime.Uptime, error) {
	msg := DBMessage{
		op:       OpGetHost,
		resp:     make(chan DBResponse),
		hostname: host,
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.one, res.err
}

func (d *Database) RemoveHost(host string) error {
	msg := DBMessage{
		op:       OpRemoveHost,
		resp:     make(chan DBResponse),
		hostname: host,
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.err
}

func (d *Database) RemovePeer(host string) error {
	msg := DBMessage{
		op:       OpRemovePeer,
		resp:     make(chan DBResponse),
		hostname: host,
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.err
}

func (d *Database) GetAllHosts() ([]uptime.Uptime, error) {
	msg := DBMessage{
		op:   OpGetAllHosts,
		resp: make(chan DBResponse),
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.all, res.err
}

func (d *Database) GetAllPeers() (map[string]bool, error) {
	msg := DBMessage{
		op:   OpGetAllPeers,
		resp: make(chan DBResponse),
	}
	d.commchan <- msg
	res := <-msg.resp
	return res.peers, res.err
}
