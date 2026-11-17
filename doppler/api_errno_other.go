//go:build !windows

package doppler

import "syscall"

var transientConnectionErrnos = []error{
	syscall.ECONNREFUSED,
	syscall.ECONNRESET,
	syscall.ECONNABORTED,
	syscall.EPIPE,
}
