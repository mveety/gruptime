package uptime

import (
	"fmt"
	"time"
)

type Uptime struct {
	Hostname string
	Time int64
}

func (u Uptime) formatUptime []string {
	var s [2]string

	s[0] = u.Hostname;
	s[1] = u.Time.String();

	return s;
}

