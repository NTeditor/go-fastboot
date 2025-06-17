package fastboot

import (
	"context"
	"fmt"

	"github.com/google/gousb"
	"github.com/nteditor/go-fastboot/fastbootErrors"
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

func (d *device) Flash(ctx context.Context, partition string, image []byte, infoHandler func([]byte)) error {
	if err := d.protocol.Download(ctx, image); err != nil {
		return err
	}
	if err := d.protocol.Send(ctx, []byte(fmt.Sprintf("flash:%s", partition))); err != nil {
		return err
	}
	for {
		if status, data, err := d.protocol.Read(ctx); err != nil {
			return err
		} else {
			switch status {
			case protocol.Status.OKAY:
				return nil
			case protocol.Status.FAIL:
				return fastbootErrors.FailedFlash
			default:
				infoHandler(data)
			}
		}
	}
}

func (d *device) GetVarAll(ctx context.Context) ([]string, error) {
	if err := d.protocol.Send(ctx, []byte("getvar:all")); err != nil {
		return nil, err
	}

	vars := []string{}
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
			return nil, fastbootErrors.FailedGetVariable
		default:
			continue
		}
	}
}

func (d *device) Close() {
	d.protocol.Close()
}
