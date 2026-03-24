package uptime

/*
Need to implement for a new platform:
	func getload() (*loadaverage, error)
	func getos() string
		This returns a string matching an OS type
	func getuptime_seconds() (time.Duration, error)
	func nusers() int
*/

import (
	"os"
	"time"
)

type Uptime struct {
	Version  int           `json:"message-version"`
	Hostname string        `json:"hostname"`
	OS       string        `json:"system"`
	Time     time.Duration `json:"uptime"`
	Load1    float64       `json:"load-average-1"`
	Load5    float64       `json:"load-average-5"`
	Load15   float64       `json:"load-average-15"`
	NUsers   uint64        `json:"users"`
	Lifetime time.Duration `json:"lifetime"`
}

type loadaverage struct {
	load1  float64
	load5  float64
	load15 float64
}

func GetUptime() (Uptime, error) {
	niluptime := Uptime{Hostname: "", Time: time.Duration(0)}
	hostname, err := os.Hostname()
	if err != nil {
		return niluptime, err
	}

	t, err := getuptime_seconds()
	if err != nil {
		return niluptime, err
	}
	l, err := getload()
	if err != nil {
		return niluptime, err
	}

	return Uptime{
		Version:  int(ProtoVersion),
		Hostname: hostname,
		OS:       getos(),
		Time:     t,
		Load1:    l.load1,
		Load5:    l.load5,
		Load15:   l.load15,
		NUsers:   uint64(nusers()),
		Lifetime: 0,
	}, nil
}

func OS2Byte(os string) byte {
	switch os {
	case "FreeBSD":
		return 1
	case "Linux":
		return 2
	case "Windows":
		return 3
	case "OpenVMS":
		return 4
	case "OpenBSD":
		return 5
	case "NetBSD":
		return 6
	case "Solaris":
		return 7
	case "Illumos":
		return 8
	case "Plan 9":
		return 9
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
	case 4:
		return "OpenVMS"
	case 5:
		return "OpenBSD"
	case 6:
		return "NetBSD"
	case 7:
		return "Solaris"
	case 8:
		return "Illumos"
	case 9:
		return "Plan 9"
	default:
		return "Unknown"
	}
}
