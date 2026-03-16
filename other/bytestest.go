package main

import (
	"fmt"

	"github.com/mveety/gruptime/internal/uptime"
)

var ProtoVersion byte = 3

// size: 0
// OS: 1
// version: 1
// uptime: version +8
// Hostname: uptime +len(Hostname)
// load1: Hostname +8
// load5: load1 +8
// load15: load5 +8
// NUsers: load15 +8

func main() {
	utime, err := uptime.GetUptime()
	if err != nil {
		panic(err)
	}
	fmt.Printf("version: %v, hostname: \"%v\", os: %v, uptime: %v, load: %v %v %v, nusers: %v, lifetime: %v\n", utime.Version, utime.Hostname, utime.OS, utime.Time, utime.Load1, utime.Load5, utime.Load15, utime.NUsers, utime.Lifetime)
	utime_bytes := utime.Bytes()
	fmt.Printf("converted: %v\n", len(utime_bytes))
	utime2, err := uptime.UptimeBuffer(utime_bytes).Uptime()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("version: %v, hostname: \"%v\", os: %v, uptime: %v, load: %v %v %v, nusers: %v, lifetime: %v\n", utime2.Version, utime2.Hostname, utime2.OS, utime2.Time, utime2.Load1, utime2.Load5, utime2.Load15, utime2.NUsers, utime2.Lifetime)
}
