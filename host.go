package fastboot

import (
	"fmt"

	"github.com/google/gousb"
	"github.com/nteditor/go-fastboot/internal/protocol"
)

type fastboot struct {
	ctx *gousb.Context
}

func (f *fastboot) ListDevices() ([]*device, error) {
	devs, err := f.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		for _, cfg := range desc.Configs {
			for _, ifc := range cfg.Interfaces {
				for _, alt := range ifc.AltSettings {
					return alt.Protocol == 0x03 && alt.Class == 0xff && alt.SubClass == 0x42
				}
			}
		}
		return true
	})

	if err != nil {
		for _, dev := range devs {
			dev.Close()
		}
		return nil, fmt.Errorf("open devices failed: %w", err)
	}

	if len(devs) == 0 {
		return []*device{}, nil
	}

	var finalDevices []*device
	for _, dev := range devs {
		intf, cleanup, err := dev.DefaultInterface()
		if err != nil {
			dev.Close()
			continue
		}
		inEndpoint, err := intf.InEndpoint(0x81)
		if err != nil {
			cleanup()
			dev.Close()
			continue
		}
		outEndpoint, err := intf.OutEndpoint(0x01)
		if err != nil {
			cleanup()
			dev.Close()
			continue
		}
		protocol := protocol.NewProtocol(inEndpoint, outEndpoint, cleanup)
		finalDevices = append(finalDevices, newDevice(dev, protocol))
	}
	return finalDevices, nil
}

func (f *fastboot) FindDeviceBySerial(serial string) (*device, error) {
	devs, err := f.ListDevices()
	if err != nil {
		return nil, err
	}
	for _, dev := range devs {
		devSerial, err := dev.Device.SerialNumber()
		if err != nil {
			return nil, err
		}
		if devSerial == serial {
			return dev, nil
		} else {
			dev.Close()
		}
	}
	return nil, nil
}

func NewHost() (*fastboot, func()) {
	ctx := gousb.NewContext()
	host := &fastboot{ctx: ctx}
	return host, host.Close
}

func (f *fastboot) Close() {
	f.ctx.Close()
}
