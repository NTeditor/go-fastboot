package fastboot

import (
	"context"
	"time"

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

func (d *device) Reboot() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

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

func (d *device) Close() {
	d.protocol.Close()
}
