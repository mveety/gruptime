package uptime

import (
	"os"
	"time"
)

type Uptime struct {
	Hostname string
	OS       string
	Time     time.Duration
	Load1    float64
	Load5    float64
	Load15   float64
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
		Hostname: hostname,
		OS:       getos(),
		Time:     t,
		Load1:    l.load1,
		Load5:    l.load5,
		Load15:   l.load15,
	}, nil
}
