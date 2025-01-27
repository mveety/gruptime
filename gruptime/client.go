package main

import (
	"encoding/gob"
	"github.com/mveety/gruptime/internal/uptime"
	"log"
	"net"
)

type TcpMessage struct {
	Uptimes []uptime.Uptime
}

func TcpConnProc(db *Database, conn net.Conn) {
	defer conn.Close()
	encoder := gob.NewEncoder(conn)
	uptimes, err := db.GetAllHosts()
	if err != nil {
		log.Fatal(err)
	}
	msg := TcpMessage{Uptimes: uptimes}
	encerr := encoder.Encode(&msg)
	if encerr != nil {
		log.Fatal(encerr)
	}
}

func TCPServer(db *Database) {
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

func TCPGetUptimes(addr string) ([]uptime.Uptime, error) {
	var msg TcpMessage
	var uptimes []uptime.Uptime
	conn, err := net.Dial("tcp", addr+":8784")
	if err != nil {
		return uptimes, err
	}
	decoder := gob.NewDecoder(conn)
	decerr := decoder.Decode(&msg)
	if decerr != nil {
		return uptimes, decerr
	}
	return msg.Uptimes, nil
}

func ReloadProc(conn net.Conn) {
	var n int
	defer conn.Close()
	peerslock.Lock()
	peers, n = readConfigfile(configfile)
	if verbose {
		log.Printf("reload: found %d hosts", n)
		if len(peers) > 0 {
			for _, s := range peers {
				log.Print(s)
			}
		}
	}
	peerslock.Unlock()
}

func ReloadServer() {
	if notcp {
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
