//go:build plan9
// +build plan9

package uptime

import (
	"os"
	"fmt"
	"strconv"
)

func crapaverage() (*loadaverage) {
	return &loadaverage{
		load1: 9.0,
		load5: 9.0,
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
		load1: float64(data[7]) / float64(ncpu*1000),
		load5: 0,
		load15: 0,
	}, nil
}
