//go:build windows
// +build windows

package uptime

func getos() string {
	return "Windows"
}
