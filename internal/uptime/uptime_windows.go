//go:build windows
// +build windows

package uptime

import (
	"syscall"
	"time"
)

var _getTickCount64 = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")

func getload() (*loadaverage, error) {
	// return something non-sensical but valid to the user
	return &loadaverage{
		load1:  -1.0,
		load5:  -1.0,
		load15: -1.0,
	}, nil
}

func getos() string {
	return "Windows"
}

func GetTickCount64() (int64, error) {
	getticks, _, err := _getTickCount64.Call()
	if errno, ok := err.(syscall.Errno); !ok || errno != 0 {
		return 0, err
	}
	return int64(getticks), nil
}

func getuptime_seconds() (time.Duration, error) {
	milli, err := GetTickCount64()
	return time.Duration(milli) * time.Millisecond, err
}

func nusers() int {
	return 0
}
