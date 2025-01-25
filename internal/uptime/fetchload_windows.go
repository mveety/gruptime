//go:build windows
// +build windows

package uptime

func getload() (*loadaverage, error) {
	// return something non-sensical but valid to the user
	return &loadaverage{
		load1:  -1.0,
		load5:  -1.0,
		load15: -1.0,
	}, nil
}
