package fastboot

import (
	"fmt"

	"github.com/google/gousb"
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
			continue
		}
		inEndpoint, err := intf.InEndpoint(0x81)
		if err != nil {
			continue
		}
		outEndpoint, err := intf.OutEndpoint(0x01)
		if err != nil {
			continue
		}
		protocol := newProtocol(inEndpoint, outEndpoint, cleanup)
		finalDevices = append(finalDevices, newDevice(dev, protocol))
	}
	return finalDevices, nil
}

func NewHost() *fastboot {
	ctx := gousb.NewContext()
	return &fastboot{
		ctx: ctx,
	}
}
