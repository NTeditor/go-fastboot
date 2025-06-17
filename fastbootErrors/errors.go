package fastbootErrors

import (
	"errors"
	"fmt"
)

var (
	ErrDeviceClose  = errors.New("connection is closed")
	ErrTimeout      = errors.New("send operation timed out")
	ErrDownload     = errors.New("failed to download file")
	ErrFlash        = errors.New("failed to flash partition")
	ErrGetVariable  = errors.New("failed to get variable")
	ErrUseGetVarAll = errors.New("use GetVarAll instead")
)

type ErrStatusFail struct {
	Data []byte
}

func (e *ErrStatusFail) Error() string {
	return fmt.Sprintf("fastboot status: \"Fail\"\nfastboot output: %s", e.Data)
}
