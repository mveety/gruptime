package uptime

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

type UptimeBuffer []byte

const (
	ProtoVersion byte = 5
)

func (u Uptime) Bytes() []byte {
	switch u.Version {
	default:
		return u.bytes3()
	case 4:
		return u.bytes4()
	case 5:
		return u.bytes5()
	}
}

func (u Uptime) bytes3() []byte {
	hostbytes := []byte(u.Hostname)
	hostlen := len(hostbytes) // size, os, uptime, loads, nusers, lifetime
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
	buf[2] = byte(u.Version)
	copy(buf[3:11], conv)
	copy(buf[11:11+hostlen], hostbytes)
	copy(buf[11+hostlen:11+hostlen+8], load1conv)
	copy(buf[11+hostlen+8:11+hostlen+16], load5conv)
	copy(buf[11+hostlen+16:11+hostlen+24], load15conv)
	copy(buf[11+hostlen+24:], nusersconv)
	return buf
}

func (u Uptime) bytes4() []byte {
	hostbytes := []byte(u.Hostname)
	hostlen := len(hostbytes) // size, os, uptime, loads, nusers, lifetime
	buf := make([]byte, hostlen+1+1+1+8+8+8+8+8+8)
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
	lifetimeconv := make([]byte, 8)
	binary.BigEndian.PutUint64(lifetimeconv, uint64(u.Lifetime))
	msglen := byte(len(buf))
	buf[0] = msglen
	buf[1] = OS2Byte(u.OS)
	buf[2] = byte(u.Version)
	copy(buf[3:11], conv)
	copy(buf[11:11+hostlen], hostbytes)
	copy(buf[11+hostlen:11+hostlen+8], load1conv)
	copy(buf[11+hostlen+8:11+hostlen+16], load5conv)
	copy(buf[11+hostlen+16:11+hostlen+24], load15conv)
	copy(buf[11+hostlen+24:11+hostlen+32], nusersconv)
	copy(buf[11+hostlen+32:], lifetimeconv)
	return buf
}

func (u Uptime) bytes5() []byte {
	hostbytes := []byte(u.Hostname)
	hostlen := len(hostbytes) // size, os, uptime, loads, nusers, lifetime
	buf := make([]byte, hostlen+1+1+1+8+8+8+8+8+8+8)
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
	lifetimeconv := make([]byte, 8)
	binary.BigEndian.PutUint64(lifetimeconv, uint64(u.Lifetime))
	issuetimeconv := make([]byte, 8)
	binary.BigEndian.PutUint64(issuetimeconv, uint64(u.Issued))
	msglen := byte(len(buf))
	buf[0] = msglen
	buf[1] = OS2Byte(u.OS)
	buf[2] = byte(u.Version)
	copy(buf[3:11], conv)
	copy(buf[11:11+hostlen], hostbytes)
	copy(buf[11+hostlen:11+hostlen+8], load1conv)
	copy(buf[11+hostlen+8:11+hostlen+16], load5conv)
	copy(buf[11+hostlen+16:11+hostlen+24], load15conv)
	copy(buf[11+hostlen+24:11+hostlen+32], nusersconv)
	copy(buf[11+hostlen+32:11+hostlen+40], lifetimeconv)
	copy(buf[11+hostlen+40:], issuetimeconv)
	return buf
}

func (msgbuf UptimeBuffer) Uptime() (Uptime, error) {
	msglen := msgbuf[0]
	if int(msglen) != len(msgbuf) {
		return Uptime{}, fmt.Errorf("message wrong size: is %d should be %d)", len(msgbuf), msglen)
	}
	switch msgbuf[2] {
	default:
		uptime, err := msgbuf.uptimev3()
		if err == nil {
			err = fmt.Errorf("message too old or too new (support 3-%d, got %d)", ProtoVersion, uptime.Version)
		}
		return uptime, err
	case 3:
		return msgbuf.uptimev3()
	case 4:
		return msgbuf.uptimev4()
	case 5:
		return msgbuf.uptimev5()
	}
}

func (msgbuf UptimeBuffer) uptimev3() (Uptime, error) {
	msglen := msgbuf[0]

	hostbuf := make([]byte, msglen-(1+1+1+8+8+8+8+8))
	hostlen := len(hostbuf)

	uptimeSeconds := int64(binary.BigEndian.Uint64(msgbuf[3:11]))
	copy(hostbuf, msgbuf[11:11+hostlen])
	hostname := string(hostbuf)
	load1bits := binary.BigEndian.Uint64(msgbuf[11+hostlen : 11+hostlen+8])
	load1 := math.Float64frombits(load1bits)
	load5bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+8 : 11+hostlen+16])
	load5 := math.Float64frombits(load5bits)
	load15bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+16 : 11+hostlen+24])
	load15 := math.Float64frombits(load15bits)
	nusers := binary.BigEndian.Uint64(msgbuf[11+hostlen+24 : 11+hostlen+32])
	return Uptime{
		Version:  3,
		Hostname: hostname,
		OS:       Byte2OS(msgbuf[1]),
		Time:     time.Duration(uptimeSeconds),
		Load1:    load1,
		Load5:    load5,
		Load15:   load15,
		NUsers:   nusers,
		Lifetime: 0,
		Issued:   0,
	}, nil
}

func (msgbuf UptimeBuffer) uptimev4() (Uptime, error) {
	msglen := msgbuf[0]

	hostbuf := make([]byte, msglen-(1+1+1+8+8+8+8+8+8))
	hostlen := len(hostbuf)

	uptimeSeconds := int64(binary.BigEndian.Uint64(msgbuf[3:11]))
	copy(hostbuf, msgbuf[11:11+hostlen])
	hostname := string(hostbuf)
	load1bits := binary.BigEndian.Uint64(msgbuf[11+hostlen : 11+hostlen+8])
	load1 := math.Float64frombits(load1bits)
	load5bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+8 : 11+hostlen+16])
	load5 := math.Float64frombits(load5bits)
	load15bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+16 : 11+hostlen+24])
	load15 := math.Float64frombits(load15bits)
	nusers := binary.BigEndian.Uint64(msgbuf[11+hostlen+24 : 11+hostlen+32])
	lifetimeSeconds := binary.BigEndian.Uint64(msgbuf[11+hostlen+32 : 11+hostlen+40])
	return Uptime{
		Version:  4,
		Hostname: hostname,
		OS:       Byte2OS(msgbuf[1]),
		Time:     time.Duration(uptimeSeconds),
		Load1:    load1,
		Load5:    load5,
		Load15:   load15,
		NUsers:   nusers,
		Lifetime: time.Duration(lifetimeSeconds),
		Issued:   0,
	}, nil
}

func (msgbuf UptimeBuffer) uptimev5() (Uptime, error) {
	msglen := msgbuf[0]

	hostbuf := make([]byte, msglen-(1+1+1+8+8+8+8+8+8+8))
	hostlen := len(hostbuf)

	uptimeSeconds := int64(binary.BigEndian.Uint64(msgbuf[3:11]))
	copy(hostbuf, msgbuf[11:11+hostlen])
	hostname := string(hostbuf)
	load1bits := binary.BigEndian.Uint64(msgbuf[11+hostlen : 11+hostlen+8])
	load1 := math.Float64frombits(load1bits)
	load5bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+8 : 11+hostlen+16])
	load5 := math.Float64frombits(load5bits)
	load15bits := binary.BigEndian.Uint64(msgbuf[11+hostlen+16 : 11+hostlen+24])
	load15 := math.Float64frombits(load15bits)
	nusers := binary.BigEndian.Uint64(msgbuf[11+hostlen+24 : 11+hostlen+32])
	lifetimeSeconds := binary.BigEndian.Uint64(msgbuf[11+hostlen+32 : 11+hostlen+40])
	issuedSeconds := binary.BigEndian.Uint64(msgbuf[11+hostlen+40 : 11+hostlen+48])
	return Uptime{
		Version:  5,
		Hostname: hostname,
		OS:       Byte2OS(msgbuf[1]),
		Time:     time.Duration(uptimeSeconds),
		Load1:    load1,
		Load5:    load5,
		Load15:   load15,
		NUsers:   nusers,
		Lifetime: time.Duration(lifetimeSeconds),
		Issued:   int64(issuedSeconds),
	}, nil
}
