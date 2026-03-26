package main

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"sort"

	"github.com/mveety/gruptime/internal/uptime"
)

const (
	ClientProtocolVersion = 4
	SUptimes              = iota
	SConfig
	SReload

	RUptimes
	RConfig
	RReload
	RError
)

type ClientError struct {
	IsNil bool
	Data  string
}

type TcpMessage struct {
	Proto   int
	Request int
}

type UptimeResponse struct {
	Uptimes []uptime.Uptime
	Peers   map[string]bool
}

type ConfigResponse struct {
	Version      string
	Rconfig      Config
	ProtoVersion int
}

type TcpResponse struct {
	Proto    int
	Response int
	Error    ClientError
	Uresp    UptimeResponse
	Cresp    ConfigResponse
}

type DownHost struct {
	Hostname string `json:"hostname"`
	Online   bool   `json:"online"`
}

type NodeStatus struct {
	uptimes   []uptime.Uptime
	uptimemap map[string]uptime.Uptime
	peers     map[string]bool
	allmap    map[string]any
	peernames []string
	hostnames []string
	allarray  []any
}

type ReloadResponse struct {
	err error
}

type ReloadMessage struct {
	resp chan ReloadResponse
}

func (e ClientError) Unwrap() error {
	if e.IsNil {
		return nil
	}
	return errors.New(e.Data)
}

func ErrorWrap(e error) ClientError {
	if e == nil {
		return ClientError{IsNil: true, Data: ""}
	}
	return ClientError{IsNil: false, Data: e.Error()}
}

func TcpConnProc(db *Database, conn net.Conn, reloadchan chan ReloadMessage) {
	defer conn.Close()
	msgdecoder := gob.NewDecoder(conn)
	var msg TcpMessage

	if err := msgdecoder.Decode(&msg); err != nil {
		log.Printf("client error: %v", err)
		return
	}
	if msg.Proto != ClientProtocolVersion {
		log.Printf("server/client versions don't match (client: %v, server: %v)", msg.Proto, ClientProtocolVersion)
		return
	}
	if runningConfig.Verbose {
		log.Printf("got client message from %v", conn.RemoteAddr())
	}
	switch msg.Request {
	default:
		log.Printf("invalid request: %v", msg.Request)
		tcpresp := TcpResponse{
			Proto:    ClientProtocolVersion,
			Response: RError,
			Error:    ErrorWrap(fmt.Errorf("invalid request: %v", msg.Request)),
			Uresp:    UptimeResponse{},
			Cresp:    ConfigResponse{},
		}
		if encerr := gob.NewEncoder(conn).Encode(&tcpresp); encerr != nil {
			log.Printf("client error: %v", encerr)
		}
	case SUptimes:
		uptimes, err := db.GetAllHosts()
		if err != nil {
			log.Fatal(err)
		}
		allpeers, err := db.GetAllPeers()
		if err != nil {
			log.Fatal(err)
		}
		tcpresp := TcpResponse{
			Proto:    ClientProtocolVersion,
			Response: RUptimes,
			Error:    ErrorWrap(nil),
			Uresp: UptimeResponse{
				Uptimes: uptimes,
				Peers:   allpeers,
			},
			Cresp: ConfigResponse{},
		}
		if encerr := gob.NewEncoder(conn).Encode(&tcpresp); encerr != nil {
			log.Printf("client error: %v", encerr)
		}
	case SConfig:
		tcpresp := TcpResponse{
			Proto:    ClientProtocolVersion,
			Response: RConfig,
			Error:    ErrorWrap(nil),
			Uresp:    UptimeResponse{},
			Cresp: ConfigResponse{
				Version:      getGitCommit(),
				Rconfig:      runningConfig,
				ProtoVersion: int(uptime.ProtoVersion),
			},
		}
		if encerr := gob.NewEncoder(conn).Encode(&tcpresp); encerr != nil {
			log.Printf("client error: %v", encerr)
		}
	case SReload:
		respchan := make(chan ReloadResponse)
		reloadchan <- ReloadMessage{resp: respchan}
		rerr := <-respchan
		tcpresp := TcpResponse{
			Proto:    ClientProtocolVersion,
			Response: RReload,
			Error:    ErrorWrap(rerr.err),
			Uresp:    UptimeResponse{},
			Cresp:    ConfigResponse{},
		}
		if encerr := gob.NewEncoder(conn).Encode(&tcpresp); encerr != nil {
			log.Printf("client error: %v", encerr)
		}
	}
}

func ClientServer(clientchan chan net.Conn, reloadchan chan ReloadMessage) {
	if runningConfig.Verbose {
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
			continue
		}
		clientchan <- conn
	}
}

func TCPGetUptimes(addr string) ([]uptime.Uptime, map[string]bool, error) {
	var uptimes []uptime.Uptime
	var errpeers map[string]bool
	conn, err := net.Dial("tcp", addr+":8784")
	if err != nil {
		return uptimes, errpeers, err
	}
	newmsg := TcpMessage{
		Proto:   ClientProtocolVersion,
		Request: SUptimes,
	}
	if err := gob.NewEncoder(conn).Encode(&newmsg); err != nil {
		return uptimes, errpeers, err
	}

	var resp TcpResponse
	if err := gob.NewDecoder(conn).Decode(&resp); err != nil {
		return uptimes, errpeers, err
	}
	if resp.Proto != ClientProtocolVersion {
		return uptimes, errpeers, errors.New("server/client mismatch")
	}
	return resp.Uresp.Uptimes, resp.Uresp.Peers, nil
}

func SendReloadMsg() error {
	conn, err := net.Dial("tcp", "127.0.0.1:8784")
	if err != nil {
		return err
	}
	defer conn.Close()

	reloadmsg := TcpMessage{
		Proto:   ClientProtocolVersion,
		Request: SReload,
	}
	if err := gob.NewEncoder(conn).Encode(&reloadmsg); err != nil {
		return err
	}

	var resp TcpResponse
	if err := gob.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}
	return resp.Error.Unwrap()
}

func getConfigData() (ConfigResponse, error) {
	conn, err := net.Dial("tcp", "127.0.0.1:8784")
	if err != nil {
		return ConfigResponse{}, err
	}
	defer conn.Close()

	confmsg := TcpMessage{
		Proto:   ClientProtocolVersion,
		Request: SConfig,
	}
	if err := gob.NewEncoder(conn).Encode(&confmsg); err != nil {
		return ConfigResponse{}, err
	}

	var resp TcpResponse
	if err := gob.NewDecoder(conn).Decode(&resp); err != nil {
		return ConfigResponse{}, err
	}
	return resp.Cresp, nil
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
		nusers = fmt.Sprintf("%d user,", u.NUsers)
	} else {
		nusers = fmt.Sprintf("%d users,", u.NUsers)
	}
	if showlifetime {
		if u.Version > 3 {
			lifetimestr = fmt.Sprintf("  (%v left)", u.Lifetime)
		} else {
			lifetimestr = "  (old version)"
		}
	}

	fmt.Printf("%-16s %-8s %s, %-9s load %.2f, %.2f, %.2f%s\n", u.Hostname, u.OS, uptime, nusers, u.Load1, u.Load5, u.Load15, lifetimestr)
}

func processNodeStatus(uptimes []uptime.Uptime, allpeers map[string]bool) *NodeStatus {
	uptimesmap := make(map[string]uptime.Uptime)
	hostnames := make([]string, len(uptimes))
	allpeersnames := make([]string, len(allpeers))
	allmap := make(map[string]any)
	allarray := make([]any, len(allpeers))
	for i, u := range uptimes {
		hostnames[i] = u.Hostname
		uptimesmap[u.Hostname] = u
	}
	i1 := 0
	for host, status := range allpeers {
		allpeersnames[i1] = host
		if status {
			allmap[host] = uptimesmap[host]
			allarray[i1] = uptimesmap[host]
		} else {
			allmap[host] = DownHost{
				Hostname: host,
				Online:   false,
			}
			allarray[i1] = allmap[host]
		}
		i1 = i1 + 1
	}

	sort.Strings(hostnames)
	sort.Strings(allpeersnames)

	return &NodeStatus{
		uptimes:   uptimes,
		uptimemap: uptimesmap,
		peers:     allpeers,
		peernames: allpeersnames,
		hostnames: hostnames,
		allmap:    allmap,
		allarray:  allarray,
	}
}

func jsonError(err error) int {
	fmt.Printf("error: unable to format as json: %v\n", err)
	return -1
}

func printNodes(nodestatus *NodeStatus, onlyliving bool, asjson bool) int {
	if onlyliving {
		if asjson {
			hostbytes, err := json.Marshal(nodestatus.hostnames)
			if err != nil {
				return jsonError(err)
			}
			fmt.Println(string(hostbytes))
			return 0
		}
		start := true
		for _, name := range nodestatus.hostnames {
			if start {
				fmt.Printf("%s", name)
				start = false
			} else {
				fmt.Printf(" %s", name)
			}
		}
	} else {
		if asjson {
			peerbytes, err := json.Marshal(nodestatus.peernames)
			if err != nil {
				return jsonError(err)
			}
			fmt.Println(string(peerbytes))
			return 0
		}
		start := true
		for _, name := range nodestatus.peernames {
			if start {
				fmt.Printf("%s", name)
				start = false
			} else {
				fmt.Printf(" %s", name)
			}
		}
	}
	fmt.Printf("\n")
	return 0
}

func printAllNodes(nodestatus *NodeStatus, onlyalive bool, asjson bool) int {
	if onlyalive {
		if asjson {
			uptimebytes, err := json.MarshalIndent(nodestatus.uptimes, "", "\t")
			if err != nil {
				return jsonError(err)
			}
			fmt.Println(string(uptimebytes))
			return 0
		}
		for _, name := range nodestatus.hostnames {
			printUptime(nodestatus.uptimemap[name])
		}
		return 0
	}

	if asjson {
		allhostbytes, err := json.MarshalIndent(nodestatus.allarray, "", "\t")
		if err != nil {
			return jsonError(err)
		}
		fmt.Println(string(allhostbytes))
		return 0
	}

	for _, name := range nodestatus.peernames {
		if nodestatus.peers[name] {
			printUptime(nodestatus.uptimemap[name])
			continue
		}
		fmt.Printf("%-16s down\n", name)
	}
	return 0
}

func printNode(nodename string, nodestatus *NodeStatus, asjson bool) int {
	status, exists := nodestatus.peers[nodename]
	if !exists {
		fmt.Printf("error: node \"%s\" not known!\n", nodename)
		return -1
	}

	if asjson {
		uptimebytes, err := json.MarshalIndent(nodestatus.allmap[nodename], "", "\t")
		if err != nil {
			return jsonError(err)
		}
		fmt.Println(string(uptimebytes))
		return 0
	}

	if status {
		printUptime(nodestatus.uptimemap[nodename])
		return 0
	}

	fmt.Printf("%-16s down\n", nodename)
	return 0
}

func clientmain(asjson bool) int {
	uptimes, allpeers, err := TCPGetUptimes("127.0.0.1")
	if err != nil {
		fmt.Printf("error: unable to connect to local daemon: %v\n", err)
		return -1
	}

	nodestatus := processNodeStatus(uptimes, allpeers)

	if printnodes {
		return printNodes(nodestatus, onlyalive, asjson)
	}

	if onlynode == "" {
		return printAllNodes(nodestatus, onlyalive, asjson)
	}

	return printNode(onlynode, nodestatus, asjson)
}
