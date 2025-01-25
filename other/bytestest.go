package main

import (
	"encoding/binary"
	"fmt"
	"github.com/mveety/gruptime/internal/uptime"
	"math"
	"time"
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
// uptime: OS +8
// Hostname: uptime +len(Hostname)
// load1: Hostname +8
// load5: load1 +8
// load15 load5 +8

func UptimeBytes(u uptime.Uptime) []byte {
	hostbytes := []byte(u.Hostname)
	hostlen := len(hostbytes) // size, os, uptime, loads
	buf := make([]byte, hostlen+1+1+8+8+8+8)
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
	msglen := byte(len(buf))
	buf[0] = msglen
	buf[1] = OS2Byte(u.OS)
	copy(buf[2:10], conv)
	copy(buf[10:10+hostlen], hostbytes)
	copy(buf[10+hostlen:10+hostlen+8], load1conv)
	copy(buf[10+hostlen+8:10+hostlen+16], load5conv)
	copy(buf[10+hostlen+16:], load15conv)
	return buf
}

func BytesUptime(msgbuf []byte) uptime.Uptime {
	msglen := msgbuf[0]
	hostbuf := make([]byte, msglen-(1+1+8+8+8+8))
	hostlen := len(hostbuf)

	uptime_seconds := int64(binary.BigEndian.Uint64(msgbuf[2:10]))
	copy(hostbuf, msgbuf[10:10+hostlen])
	hostname := string(hostbuf)
	load1bits := binary.BigEndian.Uint64(msgbuf[10+hostlen : 10+hostlen+8])
	load1 := math.Float64frombits(load1bits)
	load5bits := binary.BigEndian.Uint64(msgbuf[10+hostlen+8 : 10+hostlen+16])
	load5 := math.Float64frombits(load5bits)
	load15bits := binary.BigEndian.Uint64(msgbuf[10+hostlen+16:])
	load15 := math.Float64frombits(load15bits)
	return uptime.Uptime{
		Hostname: hostname,
		OS:       Byte2OS(msgbuf[1]),
		Time:     time.Duration(uptime_seconds),
		Load1:    load1,
		Load5:    load5,
		Load15:   load15,
	}
}

func main() {
	utime, err := uptime.GetUptime()
	if err != nil {
		panic("error getting uptime")
	}
	fmt.Printf("hostname: \"%v\", os: %v, uptime: %v, load: %v %v %v\n", utime.Hostname, utime.OS, utime.Time, utime.Load1, utime.Load5, utime.Load15)
	utime_bytes := UptimeBytes(utime)
	fmt.Printf("converted: %v\n", len(utime_bytes))
	utime2 := BytesUptime(utime_bytes)
	fmt.Printf("hostname: \"%v\", os: %v, uptime: %v, load: %v %v %v\n", utime2.Hostname, utime2.OS, utime2.Time, utime2.Load1, utime2.Load5, utime2.Load15)
}
