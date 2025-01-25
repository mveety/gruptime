//go:build linux
// +build linux

package uptime

import (
	"golang.org/x/sys/unix"
	"time"
)

func getuptime_seconds() (time.Duration, error) {
	var sysinfo unix.Sysinfo_t
	err := unix.Sysinfo(&sysinfo)
	if err != nil {
		return 0, err
	}
	return time.Duration(sysinfo.Uptime) * time.Second, nil
}
