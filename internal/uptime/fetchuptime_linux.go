// +build linux

package uptime

import (
	"fmt"
	"bufio"
	"log"
	"os"
	"string"
	"time"
	"log"
)

func getuptime_seconds() int64 {
	uptime, err := os.Open("/proc/uptime")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	uptime_reader := bufio.NewReader(uptime)
	uptime_line := uptime_reader.ReadString('\n')
	



