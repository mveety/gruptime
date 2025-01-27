package main

import (
	"encoding/binary"
	"github.com/mveety/gruptime/internal/uptime"
	"log"
	"math"
	"net"
	"time"
	"errors"
)

const (
	MulticastAddr    = "239.77.86.0:3825" // 239.M.V.0 :)
	TcpBroadcastPort = ":3826"
	ReadBuffer       = 1024 // should be big enough
)

var (
	ProtoVersion byte = 3
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
	case "Plan 9":
		return 9
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
	case 9:
		return "Plan 9"
	case 254:
		return "Unknown"
	default:
		return "invalid"
	}
}

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
	buf[1] = OS2Byte(u.OS)
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
	if msgbuf[2] < ProtoVersion {
		return uptime.Uptime{}, errors.New("protocol too old")
	}
	hostbuf := make([]byte, msglen-(1+1+1+8+8+8+8+8))
	hostlen := len(hostbuf)

	uptime_seconds := int64(binary.BigEndian.Uint64(msgbuf[3:11]))
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
		OS:       Byte2OS(msgbuf[1]),
		Time:     time.Duration(uptime_seconds),
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
		if n < int(buf[0]) {
			log.Printf("error: malformed udp message from %s: is %d should be %d", addr.String(), n, int(buf[0]))
			continue // not enough bytes for message
		}
		if verbose {
			log.Printf("got udp message from %s", addr.String())
		}
		newuptime, err := BytesUptime(buf[:n])
		if err != nil {
			log.Printf("error: udp message from %s is too old (%d < %d)", addr.String(), ProtoVersion, buf[2])
			continue
		}
		resp <- newuptime
	}
}

func udpListener(straddr string) (chan uptime.Uptime, error) {
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
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
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
	if n < int(buf[0]) {
		log.Printf("error: malformed tcp message: is %d should be %d", n, int(buf[0]))
		return
	}
	if verbose {
		log.Printf("got tcp message from %s", addr.String())
	}
	newuptime, err := BytesUptime(buf[:n])
	if err != nil {
		log.Printf("error: tcp message from %s is too old (%d < %d)", addr.String(), ProtoVersion, buf[2])
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
	ln, err := net.Listen("tcp", "0.0.0.0"+tcpport)
	if err != nil {
		return resp, err
	}
	go tcpListenerProc(ln, resp)
	return resp, nil
}

func udpBroadcasterProc(conn *net.UDPConn, trigger chan uptime.Uptime) {
	defer conn.Close()
	for {
		select {
		case newuptime := <-trigger:
			msg := UptimeBytes(newuptime)
			_, e := conn.Write(msg)
			if e != nil {
				continue // TODO: errors!
			}
		}
	}
}

func udpBroadcaster(straddr string, trigger chan uptime.Uptime) error {
	addr, err := net.ResolveUDPAddr("udp", straddr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
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
	for {
		select {
		case newuptime := <-trigger:
			for _, host := range peers {
				go tcpBroadcastWorker(host+tcpport, newuptime)
			}
		}
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
	udpchan, err := udpListener(MulticastAddr)
	if err != nil {
		log.Fatal(err)
	}
	tcpchan, err := tcpListener(TcpBroadcastPort)
	if err != nil {
		log.Fatal(err)
	}
	mechan, err := Broadcaster(MulticastAddr, TcpBroadcastPort)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case udp_neighbour_uptime := <-udpchan:
			e := d.AddHost(udp_neighbour_uptime.Hostname, udp_neighbour_uptime)
			if e != nil {
				log.Fatal(e)
			}
		case tcp_neighbour_uptime := <-tcpchan:
			e := d.AddHost(tcp_neighbour_uptime.Hostname, tcp_neighbour_uptime)
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
