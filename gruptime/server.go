package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/mveety/gruptime/internal/uptime"
	"golang.org/x/net/ipv4"
)

const (
	MulticastAddr    = "239.77.86.0" // 239.M.V.0 :)
	MulticastPort    = ":3825"
	TCPBroadcastPort = ":3826"
	ReadBuffer       = 1024 // should be big enough
)

var (
	ProtoVersion     byte = 3
	UpdateTimeout         = time.Duration(HostTimeout) * time.Second // 8 minutes
	BroadcastTimeout      = UpdateTimeout / 4
	BroadcastTTL          = 2
)

func UptimeBytes(u uptime.Uptime) []byte {
	hostbytes := []byte(u.Hostname)
	hostlen := len(hostbytes) // size, os, uptime, loads, nusers
	buf := make([]byte, hostlen+1+1+1+8+8+8+8+8)
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
	nusersconv := make([]byte, 8)
	binary.BigEndian.PutUint64(nusersconv, u.NUsers)
	msglen := byte(len(buf))
	buf[0] = msglen
	buf[1] = uptime.OS2Byte(u.OS)
	buf[2] = ProtoVersion
	copy(buf[3:11], conv)
	copy(buf[11:11+hostlen], hostbytes)
	copy(buf[11+hostlen:11+hostlen+8], load1conv)
	copy(buf[11+hostlen+8:11+hostlen+16], load5conv)
	copy(buf[11+hostlen+16:11+hostlen+24], load15conv)
	copy(buf[11+hostlen+24:], nusersconv)
	return buf
}

func BytesUptime(msgbuf []byte) (uptime.Uptime, error) {
	msglen := msgbuf[0]

	if int(msglen) != len(msgbuf) {
		return uptime.Uptime{}, fmt.Errorf("message wrong size: is %d should be %d)", len(msgbuf), msglen)
	}
	if msgbuf[2] < ProtoVersion {
		return uptime.Uptime{}, fmt.Errorf("protocol too old (%d < %d)", ProtoVersion, msgbuf[2])
	}
	hostbuf := make([]byte, msglen-(1+1+1+8+8+8+8+8))
	hostlen := len(hostbuf)

	uptimeSeconds := int64(binary.BigEndian.Uint64(msgbuf[3:11]))
	copy(hostbuf, msgbuf[11:11+hostlen])
	hostname := string(hostbuf)
	load1bits := binary.BigEndian.Uint64(msgbuf[11+hostlen : 11+hostlen+8])
	load1 := math.Float64frombits(load1bits)
	load5bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+8 : 11+hostlen+16])
	load5 := math.Float64frombits(load5bits)
	load15bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+16 : 11+hostlen+24])
	load15 := math.Float64frombits(load15bits)
	nusers := binary.BigEndian.Uint64(msgbuf[11+hostlen+24:])
	return uptime.Uptime{
		Hostname: hostname,
		OS:       uptime.Byte2OS(msgbuf[1]),
		Time:     time.Duration(uptimeSeconds),
		Load1:    load1,
		Load5:    load5,
		Load15:   load15,
		NUsers:   nusers,
	}, nil
}

func udpListenerProc(conn *net.UDPConn, resp chan uptime.Uptime) {
	defer conn.Close()
	conn.SetReadBuffer(ReadBuffer)
	for {
		buf := make([]byte, ReadBuffer)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Print("error reading UDP message")
			continue // TODO: errors!
		}
		if verbose {
			log.Printf("got udp message from %s", addr.String())
		}
		newuptime, err := BytesUptime(buf[:n])
		if err != nil {
			log.Printf("error: udp message from %s: %s", addr.String(), err)
			continue
		}
		resp <- newuptime
	}
}

func udpListener(straddr string) (chan uptime.Uptime, error) {
	var iface *net.Interface = nil
	resp := make(chan uptime.Uptime)
	if noudp {
		return resp, nil
	}
	if verbose {
		log.Print("starting udp multicast")
	}
	addr, err := net.ResolveUDPAddr("udp", straddr)
	if err != nil {
		return resp, err
	}
	if udpiface != "" {
		var err error
		iface, err = net.InterfaceByName(udpiface)
		if err != nil {
			return resp, err
		}
	}
	conn, err := net.ListenMulticastUDP("udp", iface, addr)
	if err != nil {
		return resp, err
	}
	go udpListenerProc(conn, resp)
	return resp, nil
}

func tcpListenerWorker(conn net.Conn, resp chan uptime.Uptime) {
	defer conn.Close()
	addr := conn.RemoteAddr()
	buf := make([]byte, ReadBuffer)
	n, err := conn.Read(buf)
	if err != nil {
		log.Print("error reading tcp message:", err)
		return
	}
	if verbose {
		log.Printf("got tcp message from %s", addr.String())
	}
	newuptime, err := BytesUptime(buf[:n])
	if err != nil {
		log.Printf("error: tcp message from %s: %s", addr.String(), err)
		return
	}
	resp <- newuptime
}

func tcpListenerProc(ln net.Listener, resp chan uptime.Uptime) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go tcpListenerWorker(conn, resp)
	}
}

func tcpListener(tcpport string) (chan uptime.Uptime, error) {
	resp := make(chan uptime.Uptime)
	if notcp {
		return resp, nil
	}
	if verbose {
		log.Print("starting tcp \"multicast\"")
	}
	ln, err := net.Listen("tcp", tcpbind+tcpport)
	if err != nil {
		return resp, err
	}
	go tcpListenerProc(ln, resp)
	return resp, nil
}

func udpBroadcasterProc(conn *net.UDPConn, trigger chan uptime.Uptime) {
	defer conn.Close()
	for newuptime := range trigger {
		msg := UptimeBytes(newuptime)
		_, e := conn.Write(msg)
		if e != nil {
			continue
		}
	}
}

func udpBroadcaster(straddr string, trigger chan uptime.Uptime) error {
	addr, err := net.ResolveUDPAddr("udp4", straddr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return err
	}
	if udpiface != "" {
		iface, err := net.InterfaceByName(udpiface)
		if err != nil {
			conn.Close()
			return err
		}
		packconn := ipv4.NewPacketConn(conn)
		if err := packconn.SetMulticastInterface(iface); err != nil {
			conn.Close()
			return err
		}
	}
	go udpBroadcasterProc(conn, trigger)
	return nil
}

func tcpBroadcastWorker(hostport string, u uptime.Uptime) {
	msg := UptimeBytes(u)
	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		return
	}
	defer conn.Close()
	n, e := conn.Write(msg)
	if e != nil {
		return
	}
	if n < len(msg) {
		return
	}
}

func tcpBroadcastProc(tcpport string, trigger chan uptime.Uptime) {
	for newuptime := range trigger {
		peerslock.RLock()
		for _, host := range peers {
			go tcpBroadcastWorker(host+tcpport, newuptime)
		}
		peerslock.RUnlock()
	}
}

func tcpBroadcaster(tcpport string, trigger chan uptime.Uptime) error {
	go tcpBroadcastProc(tcpport, trigger)
	return nil
}

func BroadcasterProc(mcast string, tcpport string, resp chan uptime.Uptime) {
	udptrigger := make(chan uptime.Uptime)
	tcptrigger := make(chan uptime.Uptime)
	if !noudp {
		udpe := udpBroadcaster(mcast, udptrigger)
		if udpe != nil {
			log.Fatal(udpe)
		}
	}
	if !notcp {
		tcpe := tcpBroadcaster(tcpport, tcptrigger)
		if tcpe != nil {
			log.Fatal(tcpe)
		}
	}

	startuptime, _ := uptime.GetUptime()
	resp <- startuptime
	if !noudp {
		udptrigger <- startuptime
	}
	if !notcp {
		tcptrigger <- startuptime
	}
	timer := time.NewTimer(BroadcastTimeout)
	for {
		select {
		case <-timer.C:
			newuptime, err := uptime.GetUptime()
			if err != nil {
				log.Fatal(err)
			}
			resp <- newuptime
			if !noudp {
				udptrigger <- newuptime
			}
			if !notcp {
				tcptrigger <- newuptime
			}
			timer = time.NewTimer(BroadcastTimeout)
		}
	}
}

func Broadcaster(mcastaddr string, tcpport string) (chan uptime.Uptime, error) {
	resp := make(chan uptime.Uptime)
	go BroadcasterProc(mcastaddr, tcpport, resp)
	return resp, nil
}

func Server(d *Database) {
	udpchan, err := udpListener(MulticastAddr + MulticastPort)
	if err != nil {
		log.Fatal(err)
	}
	tcpchan, err := tcpListener(TCPBroadcastPort)
	if err != nil {
		log.Fatal(err)
	}
	mechan, err := Broadcaster(MulticastAddr+MulticastPort, TCPBroadcastPort)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case udpNeighbourUptime := <-udpchan:
			e := d.AddHost(udpNeighbourUptime.Hostname, udpNeighbourUptime)
			if e != nil {
				log.Fatal(e)
			}
		case tcpNeighbourUptime := <-tcpchan:
			e := d.AddHost(tcpNeighbourUptime.Hostname, tcpNeighbourUptime)
			if e != nil {
				log.Fatal(e)
			}
		case myUptime := <-mechan:
			e := d.AddHost(myUptime.Hostname, myUptime)
			if e != nil {
				log.Fatal(e)
			}
		}
	}
}
