package fastboot

import "errors"

var FastbootErrors = struct {
	DeviceClose error
	Timeout     error
}{
	DeviceClose: errors.New("connection is closed"),
	Timeout:     errors.New("send operation timed out"),
}
