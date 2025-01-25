//go:build freebsd
// +build freebsd

package uptime

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

type load_sysctl struct {
	load  [3]uint32
	scale uint64
}

func getload() (*loadaverage, error) {
	sysctl_bytes, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		return nil, err
	}
	sysctl := *(*load_sysctl)(unsafe.Pointer(&sysctl_bytes[0]))
	return &loadaverage{
		load1:  float64(sysctl.load[0]) / float64(sysctl.scale),
		load5:  float64(sysctl.load[1]) / float64(sysctl.scale),
		load15: float64(sysctl.load[2]) / float64(sysctl.scale),
	}, nil
}
