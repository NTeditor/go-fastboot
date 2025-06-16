package fastbootErrors

import "errors"

var (
	DeviceClose    = errors.New("connection is closed")
	Timeout        = errors.New("send operation timed out")
	FailedDownload = errors.New("failed to download file")
	FailedFlash    = errors.New("failed to flash partition")
)
