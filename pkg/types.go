package fastboot

import "errors"

var Error = struct {
	VarNotFound    error
	DeviceNotFound error
	Timeout        error
}{
	VarNotFound:    errors.New("variable not found"),
	DeviceNotFound: errors.New("device not found"),
	Timeout:        errors.New("operation timeout"),
}

type FastbootResponseStatus string

var Status = struct {
	OKAY FastbootResponseStatus
	FAIL FastbootResponseStatus
	DATA FastbootResponseStatus
	INFO FastbootResponseStatus
}{
	OKAY: "OKAY",
	FAIL: "FAIL",
	DATA: "DATA",
	INFO: "INFO",
}
