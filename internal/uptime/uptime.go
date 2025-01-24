package uptime

import (
	"os"
	"time"
)

type Uptime struct {
	Hostname string
	Time     time.Duration
}

func (u Uptime) Strings() [2]string {
	var s [2]string

	s[0] = u.Hostname
	s[1] = u.Time.String()

	return s
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

	return Uptime{
		Hostname: hostname,
		Time:     time.Duration(t),
	}, nil
}
