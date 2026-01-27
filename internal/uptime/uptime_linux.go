//go:build linux
// +build linux

package uptime

//#include <utmpx.h>
// typedef struct utmpx Utmpx;
import "C"

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"time"
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

func getos() string {
	return "Linux"
}

func getuptime_seconds() (time.Duration, error) {
	var sysinfo unix.Sysinfo_t
	err := unix.Sysinfo(&sysinfo)
	if err != nil {
		return 0, err
	}
	return time.Duration(sysinfo.Uptime) * time.Second, nil
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
