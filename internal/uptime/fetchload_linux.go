//go:build linux
// +build linux

package uptime

import (
	"errors"
	"fmt"
	"os"
)

func getload() (*loadaverage, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var load loadaverage
	n, err := fmt.Fscanf(file, "%f %f %f", &load.load1, &load.load5, &load.load15)
	if err != nil || n != 3 {
		return &load, errors.New("unexpected /proc/loadavg")
	}
	return &load, nil
}
