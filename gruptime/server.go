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
		if verbose {
			log.Printf("got udp message from %s", addr.String())
		}
		newuptime, err := uptime.UptimeBuffer(buf[:n]).Uptime()
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
	newuptime, err := uptime.UptimeBuffer(buf[:n]).Uptime()
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
