package fastbooterrors

import (
	"errors"
	"fmt"
)

var (
	ErrDeviceClose  = errors.New("connection is closed")
	ErrTimeout      = errors.New("send operation timed out")
	ErrUseGetVarAll = errors.New("use GetVarAll instead")
)

type ErrStatusFail struct {
	Data []byte
}

func (e *ErrStatusFail) Error() string {
	return fmt.Sprintf("fastboot status: \"Fail\"\nfastboot output: %s", e.Data)
}
