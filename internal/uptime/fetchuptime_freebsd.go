//go:build freebsd
// +build freebsd

package uptime

/*
#include <time.h>
#include <stdint.h>
#include <errno.h>

time_t fbsd_uptime(void)
{
	struct timespec tp;
	int r;

	errno = 0;
	if(clock_gettime(CLOCK_UPTIME, &tp) < 0){
		switch(errno) {
		case EINVAL:
			return -2;
		case EPERM:
			return -3;
		default:
			return -1;
		}
	}
	return tp.tv_sec;
}
*/
import "C"
import "errors"

func getuptime_seconds() (int64, error) {
	cr := C.fbsd_uptime()
	r := int64(cr)
	switch r {
	case 0:
		return r, nil
	case -1:
		return 0, errors.New("Unknown Error")
	case -2:
		return 0, errors.New("Invalid Value (EINVAL)")
	case -3:
		return 0, errors.New("Non-superuser tried to set time (EPERM)")
	default:
		return 0, errors.New("invalid return value")
	}
}
