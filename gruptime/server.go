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

var myhostname string = ""

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
		if runningConfig.Verbose {
			if runningConfig.PrintMessages {
				log.Printf("got udp message for %s from %s: %v", newuptime.Hostname, addr.String(), newuptime)
			} else {
				log.Printf("got udp message for %s from %s", newuptime.Hostname, addr.String())
			}
		}
		resp <- newuptime
	}
}

func udpListener(straddr string) (chan uptime.Uptime, error) {
	var iface *net.Interface = nil
	resp := make(chan uptime.Uptime)
	if !runningConfig.UseUDP {
		return resp, nil
	}
	if runningConfig.Verbose {
		log.Print("listening on udp multicast")
	}
	addr, err := net.ResolveUDPAddr("udp", straddr)
	if err != nil {
		return resp, err
	}
	if runningConfig.Interface != "" {
		var err error
		iface, err = net.InterfaceByName(runningConfig.Interface)
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
	if runningConfig.Verbose {
		if runningConfig.PrintMessages {
			log.Printf("got tcp message for %s from %s: %v", newuptime.Hostname, addr.String(), newuptime)
		} else {
			log.Printf("got tcp message for %s from %s", newuptime.Hostname, addr.String())
		}
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
	if !runningConfig.UseTCP {
		return resp, nil
	}
	if runningConfig.Verbose {
		log.Print("listening on tcp \"multicast\"")
	}
	ln, err := net.Listen("tcp", runningConfig.BindAddress+tcpport)
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
	if runningConfig.Interface != "" {
		iface, err := net.InterfaceByName(runningConfig.Interface)
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
		peers := runningConfig.Peers
		for _, host := range peers {
			go tcpBroadcastWorker(host+tcpport, newuptime)
		}
	}
}

func tcpBroadcaster(tcpport string, trigger chan uptime.Uptime) error {
	go tcpBroadcastProc(tcpport, trigger)
	return nil
}

func dobroadcastall(db *Database, myhostname string, UseUDP bool, udpchan chan uptime.Uptime, UseTCP bool, tcpchan chan uptime.Uptime) {
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
		if runningConfig.Verbose {
			log.Printf("sending %v to peers", uptime)
		}
		if UseUDP {
			udpchan <- uptime
		}
		if UseTCP {
			tcpchan <- uptime
		}
	}
}

func Server(d *Database) {
	if runningConfig.Verbose {
		log.Print("starting multicast server")
	}
	udpchan, err := udpListener(MulticastAddr + MulticastPort)
	if err != nil {
		log.Fatal(err)
	}
	tcpchan, err := tcpListener(TCPBroadcastPort)
	if err != nil {
		log.Fatal(err)
	}

	udptrigger := make(chan uptime.Uptime)
	tcptrigger := make(chan uptime.Uptime)
	UseUDP := runningConfig.UseUDP
	UseTCP := runningConfig.UseTCP
	if UseUDP {
		udpe := udpBroadcaster(MulticastAddr+MulticastPort, udptrigger)
		if udpe != nil {
			log.Fatal(udpe)
		}
	}
	if UseTCP {
		tcpe := tcpBroadcaster(TCPBroadcastPort, tcptrigger)
		if tcpe != nil {
			log.Fatal(tcpe)
		}
	}

	startuptime, _ := uptime.GetUptime()
	startuptime.Lifetime = time.Duration(runningConfig.HostTimeout) * time.Second
	if e := d.AddHost(startuptime.Hostname, startuptime); e != nil {
		log.Fatal(e)
	}
	myhostname = startuptime.Hostname
	if runningConfig.Verbose {
		if runningConfig.PrintMessages {
			log.Printf("updated uptime: %v", startuptime)
		} else {
			log.Printf("update uptime")
		}
	}

	if UseUDP {
		udptrigger <- startuptime
	}
	if UseTCP {
		tcptrigger <- startuptime
	}
	timer := time.NewTimer(time.Duration(runningConfig.BroadcastInterval) * time.Second)

	for {
		select {
		case <-timer.C:
			newuptime, err := uptime.GetUptime()
			newuptime.Lifetime = time.Duration(runningConfig.HostTimeout) * time.Second
			if err != nil {
				log.Fatal(err)
			}
			if e := d.AddHost(newuptime.Hostname, newuptime); e != nil {
				log.Fatal(e)
			}
			myhostname = newuptime.Hostname
			if runningConfig.Verbose {
				if runningConfig.PrintMessages {
					log.Printf("updated uptime: %v", newuptime)
				} else {
					log.Printf("update uptime")
				}
			}
			if UseUDP {
				udptrigger <- newuptime
			}
			if UseTCP {
				tcptrigger <- newuptime
			}
			if runningConfig.Broadcast {
				dobroadcastall(d, newuptime.Hostname, UseUDP, udptrigger, UseTCP, tcptrigger)
			}
			timer = time.NewTimer(time.Duration(runningConfig.BroadcastInterval) * time.Second)
		case udpNeighbourUptime := <-udpchan:
			if udpNeighbourUptime.Hostname == myhostname {
				continue
			}
			e := d.AddHost(udpNeighbourUptime.Hostname, udpNeighbourUptime)
			if e != nil {
				log.Fatal(e)
			}
		case tcpNeighbourUptime := <-tcpchan:
			if tcpNeighbourUptime.Hostname == myhostname {
				continue
			}
			e := d.AddHost(tcpNeighbourUptime.Hostname, tcpNeighbourUptime)
			if e != nil {
				log.Fatal(e)
			}
		}
	}
}

func servermain() {
	printConfig()
	db := initUptimedb()
	go ClientServer(db)
	if !noconfig {
		go ReloadServer()
	}
	Server(db)
}
