//go:build freebsd
// +build freebsd

package uptime

import (
	"bufio"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"strconv"
	"string"
	"time"
)

func file_getuptime_seconds() (int64, error) {
	uptime_file, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer uptime_file.Close()

	uptime_line := bufio.NewReader(uptime_file).ReadString('\n')
	uptimestrs := strings.Split(uptime_line, " ")
	uptimestr, _, _ := strings.Cut(uptimestrs[0], ".")
	return strconv.ParseInt(uptimestr, 10, 64), nil
}

func getuptime_seconds() (time.Duration, error) {
	var sysinfo unix.Sysinfo_t
	err := unix.Sysinfo(&info)
	if err != nil {
		return 0, err
	}
	return time.Duration(info.Uptime) * time.Second, nil
}
