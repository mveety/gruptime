//go:build freebsd
// +build freebsd

package uptime

//#include <utmpx.h>
// typedef struct utmpx Utmpx;
import "C"

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
