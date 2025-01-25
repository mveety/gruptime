package main

import (
	"encoding/gob"
	"github.com/mveety/gruptime/internal/uptime"
	"net"
)

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
