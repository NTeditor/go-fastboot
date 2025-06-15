package fastbootErrors

import "errors"

var (
	DeviceClose = errors.New("connection is closed")
	Timeout     = errors.New("send operation timed out")
)
