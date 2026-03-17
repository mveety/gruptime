package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/mveety/gruptime/internal/uptime"
)

type Config struct {
	HostTimeout   int      `json:"timeout"`
	Broadcast     int      `json:"broadcast_interval"`
	Peers         []string `json:"peers"`
	PeerTimeout   int      `json:"peer_timeout"`
	Verbose       bool     `json:"verbose"`
	PrintMessages bool     `json:"print_messages"`
}

var (
	startserver   bool   = false
	noudp         bool   = false
	notcp         bool   = false
	configfile    string = "/usr/local/etc/gruptime.conf"
	onlynode      string = ""
	verbose       bool   = false
	verboseflag   bool   = false
	reloadconfig  bool   = false
	noreloads     bool   = false
	tcpbind       string = ""
	getversion    bool   = false
	udpiface      string = ""
	printnodes    bool   = false
	allnodes      bool   = false
	showlifetime  bool   = false
	bcastall      bool   = false
	noconfig      bool   = false
	printmessages bool   = false
	printmsgflag  bool   = false
)

func readConfigfile(file string) (Config, error) {
	confdata, err := os.ReadFile(file)
	if err != nil {
		return Config{}, err
	}

	var conf Config
	err = json.Unmarshal(confdata, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil
}

func updateConfiguration(conf Config) {
	peerslock.Lock()
	peers = conf.Peers
	peerslock.Unlock()
	if conf.HostTimeout > 0 {
		HostTimeout = time.Duration(conf.HostTimeout) * time.Second
	}
	if conf.PeerTimeout > 0 {
		PeerTimeout = time.Duration(conf.PeerTimeout) * time.Second
	}
	if conf.Broadcast > 0 {
		BroadcastTimeout = time.Duration(conf.Broadcast) * time.Second
	}
	if !verboseflag {
		verbose = conf.Verbose
	}
	if !printmsgflag {
		printmessages = conf.PrintMessages
	}
}

func printConfig() {
	if verbose {
		log.Printf("found %d peers", len(peers))
		if len(peers) > 0 {
			for _, s := range peers {
				log.Print(s)
			}
		}
		log.Printf("HostTimeout = %v", HostTimeout)
		log.Printf("PeerTimeout = %v", PeerTimeout)
		log.Printf("BroadcastTimeout = %v", BroadcastTimeout)
		log.Printf("Verbose = %v", verbose)
		log.Printf("PrintMessages = %v", printmessages)
	}
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
	peerslock = new(sync.RWMutex)

	flag.BoolVar(&startserver, "server", false, "run as gruptime server")
	flag.BoolVar(&noudp, "noudp", false, "disable udp communication (server) ")
	flag.BoolVar(&notcp, "notcp", false, "disable tcp communication (server)")
	flag.StringVar(&configfile, "config", "/usr/local/etc/gruptime.conf", "configuration file (server)")
	flag.StringVar(&onlynode, "node", "", "node to query")
	flag.BoolVar(&verboseflag, "verbose", false, "verbose output")
	flag.BoolVar(&reloadconfig, "reload", false, "reload config file (client)")
	flag.BoolVar(&noreloads, "noreloads", false, "disable config reloading (server)")
	flag.StringVar(&tcpbind, "bind", "0.0.0.0", "tcp address to bind to")
	flag.BoolVar(&getversion, "version", false, "print version and exit")
	flag.StringVar(&udpiface, "udpiface", "", "multicast on this interface")
	flag.BoolVar(&printnodes, "nodes", false, "print a list of known nodes instead of uptimes")
	flag.BoolVar(&allnodes, "all", false, "print all known nodes")
	flag.BoolVar(&showlifetime, "lifetimes", false, "show entry lifetimes")
	flag.BoolVar(&bcastall, "broadcast", false, "Send known node info to peers")
	flag.BoolVar(&noconfig, "noconfig", false, "Disable loading configuration")
	flag.BoolVar(&printmsgflag, "messages", false, "Print received messages")

	flag.Parse()

	verbose = verboseflag
	printmessages = printmsgflag

	if startserver && notcp && noudp {
		log.Fatal("error: must start either tcp or udp server")
	}

	if startserver {
		if verbose {
			log.Printf("gruptime %v", getGitCommit())
			log.Printf("protocol %v", int(uptime.ProtoVersion))
		}
		if !noconfig {
			conf, err := readConfigfile(configfile)
			if err == nil {
				updateConfiguration(conf)
			} else {
				log.Printf("unable to open configuration \"%s\": %v", configfile, err)
			}
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
