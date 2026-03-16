package main

import (
	"log"
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
	UpdateTimeout    = time.Duration(HostTimeout) * time.Second // 8 minutes
	BroadcastTimeout = UpdateTimeout / 4
	BroadcastTTL     = 2
)

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
		newuptime, err := uptime.UptimeBuffer(buf[:n]).Uptime()
		if err != nil {
			log.Printf("error: udp message from %s: %s", addr.String(), err)
			continue
		}
		if verbose {
			log.Printf("got udp message for %s from %s: %v", newuptime.Hostname, addr.String(), newuptime)
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
	newuptime, err := uptime.UptimeBuffer(buf[:n]).Uptime()
	if err != nil {
		log.Printf("error: tcp message from %s: %s", addr.String(), err)
		return
	}
	if verbose {
		log.Printf("got tcp message for %s from %s: %v", newuptime.Hostname, addr.String(), newuptime)
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
		msg := newuptime.Bytes()
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
	msg := u.Bytes()
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

func dobroadcastall(db *Database, myhostname string, udpchan chan uptime.Uptime, tcpchan chan uptime.Uptime) {
	uptimes, err := db.GetAllHosts()
	if err != nil {
		log.Printf("broadcaster: unable to get all nodes: %v", err)
		return
	}
	for _, uptime := range uptimes {
		if uptime.Hostname == myhostname {
			continue
		}
		if uptime.Version < 4 {
			continue
		}
		if verbose {
			log.Printf("sending %v to peers", uptime)
		}
		if !noudp {
			udpchan <- uptime
		}
		if !notcp {
			tcpchan <- uptime
		}
	}
}

func BroadcasterProc(db *Database, mcast string, tcpport string, resp chan uptime.Uptime) {
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
	startuptime.Lifetime = time.Duration(HostTimeout) * time.Second
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
			newuptime.Lifetime = time.Duration(HostTimeout) * time.Second
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
			if bcastall {
				go dobroadcastall(db, newuptime.Hostname, udptrigger, tcptrigger)
			}
			timer = time.NewTimer(BroadcastTimeout)
		}
	}
}

func Broadcaster(db *Database, mcastaddr string, tcpport string) (chan uptime.Uptime, error) {
	resp := make(chan uptime.Uptime)
	go BroadcasterProc(db, mcastaddr, tcpport, resp)
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
	mechan, err := Broadcaster(d, MulticastAddr+MulticastPort, TCPBroadcastPort)
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
