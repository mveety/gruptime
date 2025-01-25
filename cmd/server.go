package main

import (
	"encoding/binary"
	"github.com/mveety/gruptime/internal/uptime"
	"log"
	"math"
	"net"
	"time"
)

const (
	MulticastAddr = "239.77.86.0:3825" // 239.M.V.0 :)
	ReadBuffer    = 1024               // should be big enough
)

var (
	UpdateTimeout    = time.Duration(HostTimeout) * time.Second // 8 minutes
	BroadcastTimeout = UpdateTimeout / 4
	BroadcastTTL     = 2
)

func OS2Byte(os string) byte {
	switch os {
	case "FreeBSD":
		return 1
	case "Linux":
		return 2
	case "Windows":
		return 3
	default:
		return 254
	}
}

func Byte2OS(osbyte byte) string {
	switch osbyte {
	case 1:
		return "FreeBSD"
	case 2:
		return "Linux"
	case 3:
		return "Windows"
	case 254:
		return "Unknown"
	default:
		return "invalid"
	}
}

// size: 0
// OS: 1
// uptime: OS +8
// Hostname: uptime +len(Hostname)
// load1: Hostname +8
// load5: load1 +8
// load15 load5 +8

func UptimeBytes(u uptime.Uptime) []byte {
	hostbytes := []byte(u.Hostname)
	hostlen := len(hostbytes) // size, os, uptime, loads
	buf := make([]byte, hostlen+1+1+8+8+8+8)
	conv := make([]byte, 8)
	binary.BigEndian.PutUint64(conv, uint64(u.Time))
	load1bits := math.Float64bits(u.Load1)
	load1conv := make([]byte, 8)
	binary.BigEndian.PutUint64(load1conv, load1bits)
	load5bits := math.Float64bits(u.Load5)
	load5conv := make([]byte, 8)
	binary.BigEndian.PutUint64(load5conv, load5bits)
	load15bits := math.Float64bits(u.Load15)
	load15conv := make([]byte, 8)
	binary.BigEndian.PutUint64(load15conv, load15bits)
	msglen := byte(len(buf))
	buf[0] = msglen
	buf[1] = OS2Byte(u.OS)
	copy(buf[2:10], conv)
	copy(buf[10:10+hostlen], hostbytes)
	copy(buf[10+hostlen:10+hostlen+8], load1conv)
	copy(buf[10+hostlen+8:10+hostlen+16], load5conv)
	copy(buf[10+hostlen+16:], load15conv)
	return buf
}

func BytesUptime(msgbuf []byte) uptime.Uptime {
	msglen := msgbuf[0]
	hostbuf := make([]byte, msglen-(1+1+8+8+8+8))
	hostlen := len(hostbuf)

	uptime_seconds := int64(binary.BigEndian.Uint64(msgbuf[2:10]))
	copy(hostbuf, msgbuf[10:10+hostlen])
	hostname := string(hostbuf)
	load1bits := binary.BigEndian.Uint64(msgbuf[10+hostlen : 10+hostlen+8])
	load1 := math.Float64frombits(load1bits)
	load5bits := binary.BigEndian.Uint64(msgbuf[10+hostlen+8 : 10+hostlen+16])
	load5 := math.Float64frombits(load5bits)
	load15bits := binary.BigEndian.Uint64(msgbuf[10+hostlen+16:])
	load15 := math.Float64frombits(load15bits)
	return uptime.Uptime{
		Hostname: hostname,
		OS:       Byte2OS(msgbuf[1]),
		Time:     time.Duration(uptime_seconds),
		Load1:    load1,
		Load5:    load5,
		Load15:   load15,
	}
}

func udpListenerProc(conn *net.UDPConn, resp chan uptime.Uptime) {
	defer conn.Close()
	conn.SetReadBuffer(ReadBuffer)
	for {
		buf := make([]byte, ReadBuffer)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue // TODO: errors!
		}
		if n < int(buf[0]) {
			continue // not enough bytes for message
		}
		newuptime := BytesUptime(buf[:n])
		resp <- newuptime
	}
}

func udpListener(straddr string) (chan uptime.Uptime, error) {
	resp := make(chan uptime.Uptime)
	addr, err := net.ResolveUDPAddr("udp", straddr)
	if err != nil {
		return resp, err
	}
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return resp, err
	}
	go udpListenerProc(conn, resp)
	return resp, nil
}

func udpBroadcasterProc(conn *net.UDPConn, resp chan uptime.Uptime) {
	defer conn.Close()
	startuptime, _ := uptime.GetUptime()
	resp <- startuptime // load up the db with yourself
	timer := time.NewTimer(BroadcastTimeout)
	for {
		select {
		case <-timer.C:
			newuptime, err := uptime.GetUptime()
			if err != nil {
				log.Fatal(err)
			}
			resp <- newuptime
			msg := UptimeBytes(newuptime)
			timer = time.NewTimer(BroadcastTimeout)
			_, e := conn.Write(msg)
			if e != nil {
				continue // TODO: errors!
			}
		}
	}
}

func udpBroadcaster(straddr string) (chan uptime.Uptime, error) {
	resp := make(chan uptime.Uptime)
	addr, err := net.ResolveUDPAddr("udp", straddr)
	if err != nil {
		return resp, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return resp, err
	}
	go udpBroadcasterProc(conn, resp)
	return resp, nil
}

func UDPServer(d *Database) error {
	netchan, err := udpListener(MulticastAddr)
	if err != nil {
		log.Fatal(err)
	}
	mechan, err := udpBroadcaster(MulticastAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case neighbour_uptime := <-netchan:
			e := d.AddHost(neighbour_uptime.Hostname, neighbour_uptime)
			if e != nil {
				log.Fatal(e)
			}
		case self_uptime := <-mechan:
			e := d.AddHost(self_uptime.Hostname, self_uptime)
			if e != nil {
				log.Fatal(e)
			}
		}
	}
}
