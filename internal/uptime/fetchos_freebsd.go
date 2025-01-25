//go:build freebsd
// +build freebsd

package uptime

func getos() string {
	return "FreeBSD"
}
