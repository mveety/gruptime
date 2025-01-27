package main

import (
	"encoding/binary"
	"fmt"
	"github.com/mveety/gruptime/internal/uptime"
	"math"
	"time"
	"errors"
)

var (
	ProtoVersion byte = 3
)

func OS2Byte(os string) byte {
	switch os {
	case "FreeBSD":
		return 1
	case "Linux":
		return 2
	case "Windows":
		return 3
	default:
		return 254
	}
}

func Byte2OS(osbyte byte) string {
	switch osbyte {
	case 1:
		return "FreeBSD"
	case 2:
		return "Linux"
	case 3:
		return "Windows"
	case 254:
		return "Unknown"
	default:
		return "invalid"
	}
}

// size: 0
// OS: 1
// version: 1
// uptime: version +8
// Hostname: uptime +len(Hostname)
// load1: Hostname +8
// load5: load1 +8
// load15: load5 +8
// NUsers: load15 +8

func UptimeBytes(u uptime.Uptime) []byte {
	hostbytes := []byte(u.Hostname)
	hostlen := len(hostbytes) // size, os, uptime, loads, nusers
	buf := make([]byte, hostlen+1+1+1+8+8+8+8+8)
	conv := make([]byte, 8)
	binary.BigEndian.PutUint64(conv, uint64(u.Time))
	load1bits := math.Float64bits(u.Load1)
	load1conv := make([]byte, 8)
	binary.BigEndian.PutUint64(load1conv, load1bits)
	load5bits := math.Float64bits(u.Load5)
	load5conv := make([]byte, 8)
	binary.BigEndian.PutUint64(load5conv, load5bits)
	load15bits := math.Float64bits(u.Load15)
	load15conv := make([]byte, 8)
	binary.BigEndian.PutUint64(load15conv, load15bits)
	nusersconv := make([]byte, 8)
	binary.BigEndian.PutUint64(nusersconv, u.NUsers)
	msglen := byte(len(buf))
	buf[0] = msglen
	buf[1] = OS2Byte(u.OS)
	buf[2] = ProtoVersion
	copy(buf[3:11], conv)
	copy(buf[11:11+hostlen], hostbytes)
	copy(buf[11+hostlen:11+hostlen+8], load1conv)
	copy(buf[11+hostlen+8:11+hostlen+16], load5conv)
	copy(buf[11+hostlen+16:11+hostlen+24], load15conv)
	copy(buf[11+hostlen+24:], nusersconv)
	return buf
}

func BytesUptime(msgbuf []byte) (uptime.Uptime, error) {
	msglen := msgbuf[0]
	if msgbuf[2] < ProtoVersion {
		return uptime.Uptime{}, errors.New("protocol too old")
	}
	hostbuf := make([]byte, msglen-(1+1+1+8+8+8+8+8))
	hostlen := len(hostbuf)

	uptime_seconds := int64(binary.BigEndian.Uint64(msgbuf[3:11]))
	copy(hostbuf, msgbuf[11:11+hostlen])
	hostname := string(hostbuf)
	load1bits := binary.BigEndian.Uint64(msgbuf[11+hostlen : 11+hostlen+8])
	load1 := math.Float64frombits(load1bits)
	load5bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+8 : 11+hostlen+16])
	load5 := math.Float64frombits(load5bits)
	load15bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+16 : 11+hostlen+24])
	load15 := math.Float64frombits(load15bits)
	nusers := binary.BigEndian.Uint64(msgbuf[11+hostlen+24:])
	return uptime.Uptime{
		Hostname: hostname,
		OS:       Byte2OS(msgbuf[1]),
		Time:     time.Duration(uptime_seconds),
		Load1:    load1,
		Load5:    load5,
		Load15:   load15,
		NUsers:   nusers,
	}, nil
}

func main() {
	utime, err := uptime.GetUptime()
	if err != nil {
		panic("error getting uptime")
	}
	fmt.Printf("hostname: \"%v\", os: %v, uptime: %v, load: %v %v %v, nusers: %v\n", utime.Hostname, utime.OS, utime.Time, utime.Load1, utime.Load5, utime.Load15, utime.NUsers)
	utime_bytes := UptimeBytes(utime)
	fmt.Printf("converted: %v\n", len(utime_bytes))
	utime2, err := BytesUptime(utime_bytes)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("hostname: \"%v\", os: %v, uptime: %v, load: %v %v %v, nusers: %v\n", utime2.Hostname, utime2.OS, utime2.Time, utime2.Load1, utime2.Load5, utime2.Load15, utime.NUsers)
}
