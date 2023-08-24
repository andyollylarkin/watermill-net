package internal

import "net"

func IsTimeoutError(err error) bool {
	if err != nil {
		e, ok := err.(*net.OpError)
		if ok && e.Timeout() {
			return true
		}
	}

	return false
}
