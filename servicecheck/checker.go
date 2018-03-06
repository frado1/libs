package servicecheck

import (
	"net"
	"time"
)

type condFunc func() bool

func PingService(network string, address string) error {
	conn, err := net.DialTimeout(network, address, 50*time.Millisecond)
	if conn != nil {
		conn.Close()
	}
	return err
}

func WaitForService(network string, address string, cond condFunc, timeout time.Duration) bool {
	t := time.NewTimer(timeout)
	tick := time.NewTicker(1 * time.Second)
	defer t.Stop()
	defer tick.Stop()

	for cond() {
		select {
		case <-tick.C:
			if err := PingService(network, address); err == nil {
				return true
			}
		case <-t.C:
			break
		}
	}

	return false
}
