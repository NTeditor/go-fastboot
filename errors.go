package fastboot

import "errors"

var FastbootErrors = struct {
	deviceClose error
	timeout     error
}{
	deviceClose: errors.New("connection is closed"),
	timeout:     errors.New("send operation timed out"),
}
