//go:build freebsd
// +build freebsd

package uptime

/*
#include <time.h>
#include <stdint.h>
#include <errno.h>

typedef struct timespec Timespec;

void zero_errno(void) {
	errno = 0;
}

int get_errno(void) {
	return errno;
}
*/
import "C"

import (
	"os"
	"syscall"
	"time"
)

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
