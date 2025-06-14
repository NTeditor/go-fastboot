package fastboot

import "github.com/google/gousb"

type device struct {
	dev      *gousb.Device
	protocol *protocol
}

func newDevice(dev *gousb.Device, protocol *protocol) *device {
	return &device{
		dev:      dev,
		protocol: protocol,
	}
}

func (d *device) Reboot() error {
	return d.protocol.Send([]byte("reboot"))
}

func (d *device) Close() {
	d.protocol.Close()
}
