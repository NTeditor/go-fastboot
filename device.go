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

	resultChan := make(chan error, 1)
	go func() {
		err := d.protocol.Send(ctx, []byte("reboot"))
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			return err
		}
		d.Close()
		return nil
	case <-ctx.Done():
		return fastbootErrors.Timeout
	}
}

func (d *device) Flash(ctx context.Context, partition string, image []byte, infoHandler func([]byte)) error {

	resultChan := make(chan error, 1)
	go func() {
		if err := d.protocol.Download(ctx, image); err != nil {
			resultChan <- err
			return
		}
		if err := d.protocol.Send(ctx, []byte(fmt.Sprintf("flash:%s", partition))); err != nil {
			resultChan <- err
			return
		}
		for {
			if status, data, err := d.protocol.Read(ctx); err != nil {
				resultChan <- err
				return
			} else {
				switch status {
				case protocol.Status.OKAY:
					resultChan <- nil
					return
				case protocol.Status.FAIL:
					resultChan <- fmt.Errorf("%s", data)
					return
				default:
					infoHandler(data)
				}
			}
		}
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
		return fastbootErrors.Timeout
	}
}

func (d *device) Close() {
	d.protocol.Close()
}
