package fastboot

import "github.com/google/gousb"

func AllDevices() ([]*FastbootDevice, error) {
	context := gousb.NewContext()
	var fastbootDevices []*FastbootDevice
	devs, err := context.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		for _, cfg := range desc.Configs {
			for _, ifc := range cfg.Interfaces {
				for _, alt := range ifc.AltSettings {
					return alt.Protocol == 0x03 && alt.Class == 0xff && alt.SubClass == 0x42
				}
			}
		}
		return true
	})

	if err != nil && len(devs) == 0 {
		return nil, err
	}

	for _, dev := range devs {
		intf, done, err := dev.DefaultInterface()
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
		fastbootDevices = append(fastbootDevices, &FastbootDevice{
			Device:  dev,
			Context: context,
			In:      inEndpoint,
			Out:     outEndpoint,
			Unclaim: done,
		})
	}

	return fastbootDevices, nil
}

func FindDeviceBySerial(serial string) (*FastbootDevice, error) {
	devs, err := AllDevices()

	if err != nil {
		return &FastbootDevice{}, err
	}

	for _, dev := range devs {
		s, e := dev.Device.SerialNumber()
		if e != nil {
			continue
		}
		if serial != s {
			continue
		}
		return dev, nil
	}

	return &FastbootDevice{}, Error.DeviceNotFound
}
