//go:build illumos

package uptime

/*
#include <time.h>
#include <stdint.h>
#include <errno.h>
#include <utmpx.h>
#include <sys/loadavg.h>

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
	"errors"
	"time"
	"unsafe"
)

func getload() (*loadaverage, error) {
	var loads [3]float64

	status := C.getloadavg((*C.double)(unsafe.Pointer(&loads[0])), 3)
	if status < 3 {
		return nil, errors.New("unable to get load average")
	}
	return &loadaverage{
		load1:  loads[0],
		load5:  loads[1],
		load15: loads[2],
	}, nil
}

func getos() string {
	return "Illumos"
}

func getuptime_seconds() (time.Duration, error) {
	var ut *C.Utmpx
	var utxBootTime int64 = 0

	C.setutxent()
	for ptr := C.getutxent(); ptr != nil; ptr = C.getutxent() {
		ut = (*C.Utmpx)(ptr)
		if ut.ut_type == C.BOOT_TIME {
			utxBootTime = int64(ut.ut_tv.tv_sec)
			break
		}
	}
	C.endutxent()

	if utxBootTime == 0 {
		return time.Duration(0), errors.New("unable to find BOOT_TIME in utx")
	}

	boottime := time.Unix(utxBootTime, 0)
	return time.Since(boottime), nil
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
