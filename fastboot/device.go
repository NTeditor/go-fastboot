package fastboot

import (
	"context"
	"fmt"
	"time"

	"github.com/google/gousb"
)

type FastbootDevice struct {
	Device  *gousb.Device
	Context *gousb.Context
	In      *gousb.InEndpoint
	Out     *gousb.OutEndpoint
	Unclaim func()
}

func (d *FastbootDevice) Close() {
	d.Unclaim()
	d.Device.Close()
	d.Context.Close()
}

func (d *FastbootDevice) Send(data []byte) error {
	_, err := d.Out.Write(data)
	return err
}

func (d *FastbootDevice) GetMaxPacketSize() (int, error) {
	return d.Out.Desc.MaxPacketSize, nil
}

func (d *FastbootDevice) Recv() (FastbootResponseStatus, []byte, error) {
	var data []byte
	buf := make([]byte, d.In.Desc.MaxPacketSize)
	n, err := d.In.Read(buf)
	if err != nil {
		return Status.FAIL, []byte{}, err
	}
	data = append(data, buf[:n]...)
	var status FastbootResponseStatus
	switch string(data[:4]) {
	case "OKAY":
		status = Status.OKAY
	case "FAIL":
		status = Status.FAIL
	case "DATA":
		status = Status.DATA
	case "INFO":
		status = Status.INFO
	}
	return status, data[4:], nil
}

func (d *FastbootDevice) GetVar(variable string) (string, error) {
	err := d.Send([]byte(fmt.Sprintf("getvar:%s", variable)))
	if err != nil {
		return "", err
	}
	status, resp, err := d.Recv()
	if status == Status.FAIL {
		err = Error.VarNotFound
	}
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (d *FastbootDevice) Reboot(to string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resultChan := make(chan error, 1)

	go func() {
		err := d.Send([]byte("reboot"))
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		d.Close()
		return Error.Timeout
	}
	return nil
}

func (d *FastbootDevice) Erase(partition string) error {
	err := d.Send([]byte(fmt.Sprintf("erase:%s", partition)))
	if err != nil {
		return err
	}
	return nil
}

func (d *FastbootDevice) Flash(partition string, data []byte) error {
	err := d.Download(data)
	if err != nil {
		return err
	}

	err = d.Send([]byte(fmt.Sprintf("flash:%s", partition)))
	if err != nil {
		return err
	}

	status, data, err := d.Recv()
	switch {
	case status != Status.OKAY:
		return fmt.Errorf("failed to flash image: %s %s", status, data)
	case err != nil:
		return err
	}

	return nil
}

func (d *FastbootDevice) Download(data []byte) error {
	data_size := len(data)
	err := d.Send([]byte(fmt.Sprintf("download:%08x", data_size)))
	if err != nil {
		return err
	}

	status, _, err := d.Recv()
	switch {
	case status != Status.DATA:
		return fmt.Errorf("failed to start data phase: %s", status)
	case err != nil:
		return err
	}

	chunk_size := 0x40040

	for i := 0; i < data_size; i += chunk_size {
		end := i + chunk_size
		if end > data_size {
			end = data_size
		}
		err := d.Send(data[i:end])
		if err != nil {
			return err
		}
	}
	status, data, err = d.Recv()
	switch {
	case status != Status.OKAY:
		return fmt.Errorf("failed to finish data phase: %s %s", status, data)
	case err != nil:
		return err
	}
	return nil
}
