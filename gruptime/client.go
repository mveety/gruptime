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
