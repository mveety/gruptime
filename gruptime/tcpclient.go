package main

import (
	"encoding/gob"
	"github.com/mveety/gruptime/internal/uptime"
	"net"
)

func TCPGetUptimes(addr string) ([]uptime.Uptime, error) {
	var uptimes []uptime.Uptime
	conn, err := net.Dial("tcp", addr+":8784")
	if err != nil {
		return uptimes, err
	}
	decoder := gob.NewDecoder(conn)
	decerr := decoder.Decode(&uptimes)
	if decerr != nil {
		return uptimes, decerr
	}
	return uptimes, nil
}
