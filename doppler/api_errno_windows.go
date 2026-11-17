//go:build windows

package doppler

import "syscall"

// Windows reports Winsock codes, which never match the syscall.ECONN* values Go defines for
// portability.
var transientConnectionErrnos = []error{
	syscall.Errno(10061), // WSAECONNREFUSED, not defined in syscall
	syscall.WSAECONNRESET,
	syscall.WSAECONNABORTED,
}
