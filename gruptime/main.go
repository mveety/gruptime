package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime/debug"
	"sync"

	"github.com/mveety/gruptime/internal/uptime"
)

var (
	startserver  bool   = false
	noudp        bool   = false
	notcp        bool   = false
	configfile   string = "/usr/local/etc/gruptime.conf"
	peers        []string
	peerslock    *sync.RWMutex
	onlynode     string = ""
	verbose      bool   = false
	reloadconfig bool   = false
	noreloads    bool   = false
	tcpbind      string = ""
	getversion   bool   = false
	udpiface     string = ""
	printnodes   bool   = false
	allnodes     bool   = false
	showlifetime bool   = false
)

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
		if allnodes {
			start := true
			for k := range allpeers {
				if start {
					fmt.Printf("%s", k)
					start = false
				} else {
					fmt.Printf(" %s", k)
				}
			}
		} else {
			start := true
			for _, u := range uptimes {
				if start {
					fmt.Printf("%s", u.Hostname)
					start = false
				} else {
					fmt.Printf(" %s", u.Hostname)
				}
			}
		}
		fmt.Printf("\n")
		return
	}

	if onlynode == "" {
		if allnodes {
			for k := range allpeers {
				if allpeers[k] {
					printUptime(uptimesmap[k])
				} else {
					fmt.Printf("%-16s down\n", k)
				}
			}
		} else {
			for _, u := range uptimes {
				printUptime(u)
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

func servermain() {
	db := initUptimedb()
	if verbose {
		log.Print("starting tcp server")
	}
	go TCPServer(db)
	if !noreloads {
		go ReloadServer()
	}
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

func getGitCommit() string {
	if buildinfo, ok := debug.ReadBuildInfo(); ok {
		for _, s := range buildinfo.Settings {
			if s.Key == "vcs.revision" {
				return s.Value
			}
		}
	}
	return "(unknown)"
}

func main() {
	var n int
	peerslock = new(sync.RWMutex)

	flag.BoolVar(&startserver, "server", false, "run as gruptime server")
	flag.BoolVar(&noudp, "noudp", false, "disable udp communication (server) ")
	flag.BoolVar(&notcp, "notcp", false, "disable tcp communication (server)")
	flag.StringVar(&configfile, "config", "/usr/local/etc/gruptime.conf", "configuration file (server)")
	flag.StringVar(&onlynode, "node", "", "node to query")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.BoolVar(&reloadconfig, "reload", false, "reload config file (client)")
	flag.BoolVar(&noreloads, "noreloads", false, "disable config reloading (server)")
	flag.StringVar(&tcpbind, "bind", "0.0.0.0", "tcp address to bind to")
	flag.BoolVar(&getversion, "version", false, "print version and exit")
	flag.StringVar(&udpiface, "udpiface", "", "multicast on this interface")
	flag.BoolVar(&printnodes, "nodes", false, "print a list of known nodes instead of uptimes")
	flag.BoolVar(&allnodes, "all", false, "print all known nodes")
	flag.BoolVar(&showlifetime, "lifetimes", false, "show entry lifetimes")

	flag.Parse()

	if startserver && notcp && noudp {
		log.Fatal("error: must start either tcp or udp server")
	}

	if startserver {
		if verbose {
			log.Printf("gruptime %v", getGitCommit())
			log.Printf("protocol %v", int(uptime.ProtoVersion))
		}
		if !notcp {
			peerslock.Lock()
			peers, n = readConfigfile(configfile)
			if verbose {
				log.Printf("found %d hosts", n)
				if len(peers) > 0 {
					for _, s := range peers {
						log.Print(s)
					}
				}
			}
			peerslock.Unlock()
		}
		servermain()
	} else {
		if getversion {
			fmt.Printf("gruptime %v\ngruptime protocol %v\n", getGitCommit(), int(uptime.ProtoVersion))
			os.Exit(0)
		}
		if reloadconfig {
			SendReloadMsg()
		} else {
			clientmain()
		}
	}
}
