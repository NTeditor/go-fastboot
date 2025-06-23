package fastboot

import (
	"context"
	"fmt"

	"github.com/google/gousb"
	"github.com/nteditor/go-fastboot/fastbooterrors"
	"github.com/nteditor/go-fastboot/internal/protocol"
)

type device struct {
	Device   *gousb.Device
	protocol *protocol.Protocol
}

func newDevice(dev *gousb.Device, protocol *protocol.Protocol) *device {
	return &device{
		Device:   dev,
		protocol: protocol,
	}
}

func (d *device) Reboot(ctx context.Context) error {
	err := d.protocol.Send(ctx, []byte("reboot"))
	if err != nil {
		return err
	}
	d.Close()
	return nil
}

func (d *device) GetVarAll(ctx context.Context) ([]string, error) {
	if err := d.protocol.Send(ctx, []byte("getvar:all")); err != nil {
		return nil, err
	}

	var vars []string
	for {
		status, data, err := d.protocol.Read(ctx)
		if err != nil {
			return nil, err
		}
		switch status {
		case protocol.Status.OKAY:
			return vars, nil
		case protocol.Status.DATA, protocol.Status.INFO:
			vars = append(vars, string(data))
		case protocol.Status.FAIL:
			return vars, &fastbooterrors.ErrStatusFail{Data: data}
		default:
			continue
		}
	}
}

func (d *device) GetVar(ctx context.Context, variable string) (string, error) {
	if variable == "all" {
		return "", fastbooterrors.ErrUseGetVarAll
	}

	if err := d.protocol.Send(ctx, []byte(fmt.Sprintf("getvar:%s", variable))); err != nil {
		return "", err
	}

	for {
		status, data, err := d.protocol.Read(ctx)
		if err != nil {
			return "", err
		}
		switch status {
		case protocol.Status.OKAY:
			return string(data), nil
		case protocol.Status.FAIL:
			return "", &fastbooterrors.ErrStatusFail{Data: data}
		default:
			continue
		}
	}
}

func (d *device) Close() {
	if !d.protocol.IsClosed {
		d.protocol.IsClosed = true
		d.protocol.Cleanup()
		d.Device.Close()
	}
}
