//go:build plan9
// +build plan9

package uptime

import (
//	"errors"
	"fmt"
	"os"
	"time"
)

func getuptime_seconds() (time.Duration, error) {
	file, err := os.Open("/dev/time")
	if err != nil {
		return time.Duration(0), err
	}
	defer file.Close()
	var timedata [5]int64
	n, err := fmt.Fscanf(file, "%d %d %d %d %d", &timedata[0], &timedata[1], &timedata[2], &timedata[3], &timedata[4])
	if err != nil || n < 5 {
		return time.Duration(0), err
	}

	return time.Duration(timedata[2]/timedata[3]) * time.Second, nil
}
