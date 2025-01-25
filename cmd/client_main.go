package main

import (
	"fmt"
	"os"
	"github.com/mveety/gruptime/internal/uptime"
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
