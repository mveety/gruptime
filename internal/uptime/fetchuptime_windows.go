//go:build windows
// +build windows

package uptime

import (
	"syscall"
	"time"
)

var _getTickCount64 = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")

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
