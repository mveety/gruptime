//go:build freebsd
// +build freebsd

package uptime

/*
#include <time.h>
#include <stdint.h>
#include <errno.h>
#include <utmpx.h>

typedef struct timespec Timespec;
typedef struct utmpx Utmpx;

void zero_errno(void) {
	errno = 0;
}

int get_errno(void) {
	return errno;
}
*/
import "C"

import (
	"golang.org/x/sys/unix"
	"os"
	"syscall"
	"time"
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

func getos() string {
	return "FreeBSD"
}

func getuptime_seconds() (time.Duration, error) {
	var ts C.Timespec

	C.zero_errno()
	if C.clock_gettime(C.CLOCK_UPTIME, &ts) < 0 {
		errno := syscall.Errno(C.get_errno())
		switch errno {
		case syscall.EINVAL:
			return time.Duration(0), os.ErrInvalid
		case syscall.EPERM:
			return time.Duration(0), os.ErrPermission
		}
	}
	return time.Duration(int64(ts.tv_sec)) * time.Second, nil
}

func fetchusers() []string {
	found := make(map[string]int)
	var ut *C.Utmpx

	C.setutxent()

	for ptr := C.getutxent(); ptr != nil; ptr = C.getutxent() {
		ut = (*C.Utmpx)(ptr)
		if ut.ut_type == C.USER_PROCESS {
			ut_user := ut.ut_user[0:]
			found[C.GoString(&ut_user[0])] = 1
		}
	}

	C.endutxent()

	users := make([]string, len(found))
	i := 0
	for k := range found {
		users[i] = k
		i++
	}
	return users
}

func nusers() int {
	users := fetchusers()
	return len(users)
}
