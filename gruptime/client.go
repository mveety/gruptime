package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"

	"github.com/mveety/gruptime/internal/uptime"
)

const (
	ClientProtocolVersion = 3
)

type TcpMessage struct {
	Proto   int
	Uptimes []uptime.Uptime
	Peers   map[string]bool
}

func TcpConnProc(db *Database, conn net.Conn) {
	defer conn.Close()
	encoder := gob.NewEncoder(conn)
	uptimes, err := db.GetAllHosts()
	if err != nil {
		log.Fatal(err)
	}
	allpeers, err := db.GetAllPeers()
	if err != nil {
		log.Fatal(err)
	}
	msg := TcpMessage{Proto: ClientProtocolVersion, Uptimes: uptimes, Peers: allpeers}
	encerr := encoder.Encode(&msg)
	if encerr != nil {
		log.Fatal(encerr)
	}
}

func ClientServer(db *Database) {
	if verbose {
		log.Print("starting client connection server")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:8784") // UPTI
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
		}
		go TcpConnProc(db, conn)
	}
}

func TCPGetUptimes(addr string) ([]uptime.Uptime, map[string]bool, error) {
	var msg TcpMessage
	var uptimes []uptime.Uptime
	var errpeers map[string]bool
	conn, err := net.Dial("tcp", addr+":8784")
	if err != nil {
		return uptimes, errpeers, err
	}
	decoder := gob.NewDecoder(conn)
	decerr := decoder.Decode(&msg)
	if decerr != nil {
		return uptimes, errpeers, decerr
	}
	if msg.Proto != ClientProtocolVersion {
		return uptimes, errpeers, errors.New("server/client mismatch")
	}
	return msg.Uptimes, msg.Peers, nil
}

func ReloadProc(conn net.Conn) {
	defer conn.Close()
	conf, err := readConfigfile(configfile)
	if err != nil {
		log.Printf("unable to read config file \"%s\": %v", configfile, err)
		return
	}
	if verbose {
		log.Printf("reloading config file \"%s\"", configfile)
	}
	updateConfiguration(conf)
	printConfig()
}

func ReloadServer() {
	if noreloads {
		log.Print("configuration reloading disabled")
		return
	}
	if verbose {
		log.Print("starting reload server")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:8785")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
		}
		go ReloadProc(conn)
	}
}

func SendReloadMsg() error {
	conn, err := net.Dial("tcp", "127.0.0.1:8785")
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func printUptime(u uptime.Uptime) {
	uptimeDays := int(math.Floor(u.Time.Hours()) / 24)
	uptimeHours := int(u.Time.Hours()) % 24
	uptimeMinutes := int(u.Time.Minutes()) % 60
	uptimeSeconds := int(u.Time.Seconds()) % 60
	var uptime string
	var nusers string
	lifetimestr := ""
	if uptimeDays < 1 {
		if uptimeHours < 1 {
			if uptimeMinutes < 1 {
				uptime = fmt.Sprintf("%d seconds", uptimeSeconds)
			} else {
				uptime = fmt.Sprintf("%d minutes", uptimeMinutes)
			}
		} else {
			uptime = fmt.Sprintf("%2d:%02d", uptimeHours, uptimeMinutes)
		}
	} else {
		uptime = fmt.Sprintf("%d+%02d:%02d", uptimeDays, uptimeHours, uptimeMinutes)
	}
	if u.NUsers == 1 {
		nusers = fmt.Sprintf("%d user", u.NUsers)
	} else {
		nusers = fmt.Sprintf("%d users", u.NUsers)
	}
	if showlifetime {
		if u.Version > 3 {
			lifetimestr = fmt.Sprintf("  (%v left)", u.Lifetime)
		} else {
			lifetimestr = "  (old version)"
		}
	}

	fmt.Printf("%-16s %-8s %s, %s, load %.2f, %.2f, %.2f%s\n", u.Hostname, u.OS, uptime, nusers, u.Load1, u.Load5, u.Load15, lifetimestr)
}

func clientmain() {
	defer os.Exit(0)
	uptimes, allpeers, err := TCPGetUptimes("127.0.0.1")
	if err != nil {
		fmt.Printf("error: unable to connect to local daemon: %v\n", err)
		return
	}

	uptimesmap := make(map[string]uptime.Uptime)
	for _, u := range uptimes {
		uptimesmap[u.Hostname] = u
	}

	if printnodes {
		if onlyalive {
			start := true
			for _, u := range uptimes {
				if start {
					fmt.Printf("%s", u.Hostname)
					start = false
				} else {
					fmt.Printf(" %s", u.Hostname)
				}
			}
		} else {
			start := true
			for k := range allpeers {
				if start {
					fmt.Printf("%s", k)
					start = false
				} else {
					fmt.Printf(" %s", k)
				}
			}
		}
		fmt.Printf("\n")
		return
	}

	if onlynode == "" {
		if onlyalive {
			for _, u := range uptimes {
				printUptime(u)
			}
		} else {
			for k := range allpeers {
				if allpeers[k] {
					printUptime(uptimesmap[k])
				} else {
					fmt.Printf("%-16s down\n", k)
				}
			}
		}
		return
	}

	status, exists := allpeers[onlynode]
	if !exists {
		fmt.Printf("error: node \"%s\" not known!\n", onlynode)
		os.Exit(-1)
	}
	if status {
		printUptime(uptimesmap[onlynode])
	} else {
		fmt.Printf("%-16s down\n", onlynode)
	}
}
