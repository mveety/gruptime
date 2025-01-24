//go:build freebsd
// +build freebsd

package uptime

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"string"
	"time"
)

func getuptime_seconds() (int64, error) {
	uptime_file, err := os.Open("/proc/uptime")
	if err != nil {
		return nil, err
	}
	defer uptime_file.Close()

	uptime_line := bufio.NewReader(uptime_file).ReadString('\n')
	uptimestrs := strings.Split(uptime_line, " ")
	uptimestr, _, _ := strings.Cut(uptimestrs[0], ".")
	return strconv.ParseInt(uptimestr, 10, 64), nil
}
