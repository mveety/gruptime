//go:build plan9
// +build plan9

package uptime

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func crapaverage() *loadaverage {
	return &loadaverage{
		load1:  9.0,
		load5:  9.0,
		load15: 9.0,
	}
}

func getload() (*loadaverage, error) {
	ncpu, err := strconv.Atoi(os.Getenv("NPROC"))
	if err != nil {
		return nil, err
	}
	file, err := os.Open("/dev/sysstat")
	if err != nil {
		// if we can't open /dev/sysstat just quit and
		// give a nonsense load.
		return crapaverage(), nil
	}
	defer file.Close()
	var data [10]int64
	n, err := fmt.Fscanf(file, "%d %d %d %d %d %d %d %d %d %d\n", &data[0], &data[1], &data[2], &data[3], &data[4], &data[5], &data[6], &data[7], &data[8], &data[9])
	if err != nil || n < 10 {
		return crapaverage(), err
	}
	return &loadaverage{
		load1:  float64(data[7]) / float64(ncpu*1000),
		load5:  0,
		load15: 0,
	}, nil
}

func getos() string {
	return "Plan 9"
}

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

func nusers() int {
	return 0
}
