//go:build windows
// +build windows

package uptime

import (
	"syscall"
	"time"
)

var _getTickCount64 = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")

func GetTickCount64() (int64, error) {
	getticks, _, err := _getTickCount64()
	if errno, ok := err.(syscall.Errno); ok != nil || errno != 0 {
		return time.Duration(0), err
	}
	return getticks, nil
}

func getuptime_seconds() (time.Duration, error) {
	milli, err := GetTickCount64()
	return time.Duration(milli) * time.Millisecond, err
}
