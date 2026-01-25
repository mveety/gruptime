//go:build plan9
// +build plan9

package uptime

func getload() (*loadaverage, error) {
	return &loadaverage{
		load1:  9.0,
		load5:  9.0,
		load15: 9.0,
	}, nil
}
