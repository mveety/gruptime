package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/mveety/gruptime/internal/uptime"
)

type Config struct {
	HostTimeout       int      `json:"timeout"`
	BroadcastInterval int      `json:"broadcast_interval"`
	Peers             []string `json:"peers"`
	PeerTimeout       int      `json:"peer_timeout"`
	Verbose           bool     `json:"verbose"`
	PrintMessages     bool     `json:"print_messages"`
	Broadcast         bool     `json:"broadcast"`
	UseTCP            bool     `json:"use_tcp"`
	UseUDP            bool     `json:"use_udp"`
	BindAddress       string   `json:"bind_address"`
	Interface         string   `json:"interface"`
}

var defaultConfig Config = Config{
	HostTimeout:       480,
	PeerTimeout:       86000,
	BroadcastInterval: 240,
	Verbose:           false,
	PrintMessages:     false,
	Broadcast:         true,
	UseTCP:            false,
	UseUDP:            true,
	BindAddress:       "0.0.0.0",
	Interface:         "",
}

var runningConfig Config

var (
	startserver   bool   = false
	configfile    string = ""
	confargdef    string = ""
	defconfigfile string = "/usr/local/etc/gruptime.conf"
	altconfigfile string = "/etc/gruptime.conf"
	onlynode      string = ""
	reloadconfig  bool   = false
	noreloads     bool   = false
	getversion    bool   = false
	printnodes    bool   = false
	onlyalive     bool   = false
	showlifetime  bool   = false
	noconfig      bool   = false
	verbose       bool   = false
)

func readConfigfile(file string) (Config, error) {
	confdata, err := os.ReadFile(file)
	if err != nil {
		return Config{}, err
	}

	conf := defaultConfig
	err = json.Unmarshal(confdata, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil
}

func updateConfiguration(conf Config) {
	runningConfig = conf
	if verbose {
		runningConfig.Verbose = verbose
	}
}

func printConfig() {
	if runningConfig.Verbose {
		log.Printf("found %d peers", len(runningConfig.Peers))
		if len(runningConfig.Peers) > 0 {
			for _, s := range runningConfig.Peers {
				log.Print(s)
			}
		}
		log.Printf("HostTimeout = %v", time.Duration(runningConfig.HostTimeout)*time.Second)
		log.Printf("PeerTimeout = %v", time.Duration(runningConfig.PeerTimeout)*time.Second)
		log.Printf("Broadcast = %v", runningConfig.Broadcast)
		log.Printf("BroadcastInterval = %v", time.Duration(runningConfig.BroadcastInterval)*time.Second)
		log.Printf("Verbose = %v", runningConfig.Verbose)
		log.Printf("PrintMessages = %v", runningConfig.PrintMessages)
		log.Printf("UseTCP = %v", runningConfig.UseTCP)
		log.Printf("UseUDP = %v", runningConfig.UseUDP)
		log.Printf("BindAddress = %v", runningConfig.BindAddress)
		log.Printf("Interface = %v", runningConfig.Interface)
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

func setdefaultconfigfile() {
	_, err := os.Stat(defconfigfile)
	if err == nil {
		confargdef = defconfigfile
		return
	}
	_, err = os.Stat(altconfigfile)
	if err == nil {
		confargdef = altconfigfile
		return
	}
	confargdef = defconfigfile
}

func main() {
	setdefaultconfigfile()

	flag.BoolVar(&startserver, "server", false, "run as gruptime server")
	flag.StringVar(&configfile, "config", confargdef, "configuration file (server)")
	flag.StringVar(&onlynode, "node", "", "node to query")
	flag.BoolVar(&reloadconfig, "reload", false, "reload config file (client)")
	flag.BoolVar(&noreloads, "noreloads", false, "disable config reloading (server)")
	flag.BoolVar(&getversion, "version", false, "print version and exit")
	flag.BoolVar(&printnodes, "nodes", false, "print a list of known nodes instead of uptimes")
	flag.BoolVar(&onlyalive, "alive", false, "only print living nodes")
	flag.BoolVar(&showlifetime, "lifetimes", false, "show entry lifetimes")
	flag.BoolVar(&noconfig, "noconfig", false, "Disable loading configuration")
	flag.BoolVar(&verbose, "verbose", false, "print debugging messages")

	flag.Parse()

	runningConfig = defaultConfig

	if startserver {
		if runningConfig.Verbose {
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
		if !runningConfig.UseTCP && !runningConfig.UseUDP {
			log.Fatal("must set UseUDP or UseTCP")
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
