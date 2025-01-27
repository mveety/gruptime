package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/mveety/gruptime/internal/uptime"
	"log"
	"math"
	"os"
)

var (
	startserver bool   = false
	noudp       bool   = false
	notcp       bool   = false
	configfile  string = "/usr/local/etc/gruptime.conf"
	peers       []string
	onlynode    string = ""
	verbose     bool   = false
)

func printUptime(u uptime.Uptime) {
	uptime_days := int(math.Floor(u.Time.Hours()) / 24)
	uptime_hours := int(u.Time.Hours()) % 24
	uptime_minutes := int(u.Time.Minutes()) % 60
	uptime_seconds := int(u.Time.Seconds()) % 60
	var uptime string
	var nusers string
	if uptime_days < 1 {
		if uptime_hours < 1 {
			if uptime_minutes < 1 {
				uptime = fmt.Sprintf("%d seconds", uptime_seconds)
			} else {
				uptime = fmt.Sprintf("%d minutes", uptime_minutes)
			}
		} else {
			uptime = fmt.Sprintf("%2d:%02d", uptime_hours, uptime_minutes)
		}
	} else {
		uptime = fmt.Sprintf("%d+%02d:%02d", uptime_days, uptime_hours, uptime_minutes)
	}
	if u.NUsers == 1 {
		nusers = fmt.Sprintf("%d user", u.NUsers)
	} else {
		nusers = fmt.Sprintf("%d users", u.NUsers)
	}
	fmt.Printf("%-16s %-8s %s, %s, load %.2f, %.2f, %.2f\n", u.Hostname, u.OS, uptime, nusers, u.Load1, u.Load5, u.Load15)
}

func clientmain() {
	uptimes, err := TCPGetUptimes("127.0.0.1")
	if err != nil {
		fmt.Printf("error: unable to connect to local daemon: %v\n", err)
	}

	if onlynode == "" {
		for _, u := range uptimes {
			printUptime(u)
		}
	} else {
		for _, u := range uptimes {
			if u.Hostname == onlynode {
				printUptime(u)
			}
		}
	}
	os.Exit(0)
}

func servermain() {
	db := initUptimedb()
	if verbose {
		log.Print("starting tcp server")
	}
	go TCPServer(db)
	if verbose {
		log.Print("starting multicast server")
	}
	Server(db)
}

func readConfigfile(file string) ([]string, int) {
	var tmp []string
	fd, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	i := 0
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		tmp = append(tmp, line)
		i++
	}
	return tmp, i
}

func main() {
	var n int
	flag.BoolVar(&startserver, "daemon", false, "run as gruptime daemon")
	flag.BoolVar(&noudp, "noudp", false, "disable udp communication")
	flag.BoolVar(&notcp, "notcp", false, "disable tcp communication")
	flag.StringVar(&configfile, "config", "/usr/local/etc/gcruptime.conf", "configuration file")
	flag.StringVar(&onlynode, "node", "", "node to query")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")

	flag.Parse()
	if startserver && notcp && noudp {
		log.Fatal("error: must start either tcp or udp server")
	}

	if startserver {
		if !notcp {
			peers, n = readConfigfile(configfile)
			if verbose {
				log.Printf("found %d hosts", n)
				if len(peers) > 0 {
					for _, s := range peers {
						log.Print(s)
					}
				}
			}
		}
		servermain()
	} else {
		clientmain()
	}
}
