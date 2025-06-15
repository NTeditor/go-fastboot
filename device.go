package fastboot

import (
	"context"
	"time"

	"github.com/google/gousb"
)

type device struct {
	Device   *gousb.Device
	protocol *protocol
}

func newDevice(dev *gousb.Device, protocol *protocol) *device {
	return &device{
		Device:   dev,
		protocol: protocol,
	}
}

func (d *device) Reboot() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := d.protocol.Send(ctx, []byte("reboot")); err == nil {
		d.Close()
		return nil
	} else {
		return err
	}
}

func (d *device) Close() {
	d.protocol.Close()
}
