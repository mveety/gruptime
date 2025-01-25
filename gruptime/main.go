package main

import (
	"fmt"
	"github.com/mveety/gruptime/internal/uptime"
	"os"
	"log"
	"flag"
)

func printUptime(u uptime.Uptime) {
	fmt.Printf("%-16s %-8s %v, load: %2f %2f %2f\n", u.Hostname, u.OS, u.Time, u.Load1, u.Load5, u.Load15)
}

func clientmain() {
	uptimes, err := TCPGetUptimes("127.0.0.1")
	if err != nil {
		fmt.Printf("error: unable to connect to local daemon: %v\n", err)
	}

	for _, u := range uptimes {
		printUptime(u)
	}
	os.Exit(0)
}

func servermain() {
	db := initUptimedb()
	log.Print("starting tcp server")
	go TCPServer(db)
	log.Print("starting udp multicast server")
	UDPServer(db)
}

func main() {
	startserver := flag.Bool("daemon", false, "run as gruptime daemon")

	flag.Parse()
	if *startserver {
		servermain()
	} else {
		clientmain()
	}
}

