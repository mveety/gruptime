package main

import (
	"flag"
	"fmt"
	"github.com/mveety/gruptime/internal/uptime"
	"log"
	"os"
	"bufio"
)

var (
	startserver bool   = false
	noudp       bool   = false
	notcp       bool   = false
	configfile  string = "/usr/local/etc/gruptime.conf"
	peers       []string
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
	log.Print("starting multicast server")
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

	flag.Parse()
	if startserver && notcp && noudp {
		log.Fatal("error: must start either tcp or udp server");
	}

	if startserver {
		peers, n = readConfigfile(configfile)
		log.Printf("found %d hosts", n)
		if len(peers) > 0 {
			for _, s := range peers {
				log.Print(s)
			}
		}
		servermain()
	} else {
		clientmain()
	}
}
