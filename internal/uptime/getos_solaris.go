//go:build solaris && !illumos

package uptime

func getos() string {
	return "Solaris"
}
