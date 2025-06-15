package fastboot

import (
	"context"

	"github.com/google/gousb"
)

type FastbootStatus int

const (
	OKAY FastbootStatus = iota
	FAIL
	DATA
	INFO
)

func status(data []byte) FastbootStatus {
	switch string(data[:4]) {
	case "OKAY":
		return OKAY
	case "INFO":
		return INFO
	case "DATA":
		return DATA
	default:
		return FAIL
	}

}

type protocol struct {
	inEndpoint  *gousb.InEndpoint
	outEndpoint *gousb.OutEndpoint
	cleanup     func()
	IsClosed    bool
}

func newProtocol(
	inEndpoint *gousb.InEndpoint,
	outEndpoint *gousb.OutEndpoint,
	cleanup func(),
) *protocol {
	return &protocol{
		inEndpoint:  inEndpoint,
		outEndpoint: outEndpoint,
		cleanup:     cleanup,
	}
}

func (p *protocol) Send(ctx context.Context, data []byte) error {
	if p.IsClosed {
		return FastbootErrors.DeviceClose
	}
	_, err := p.outEndpoint.WriteContext(ctx, data)
	return err
}

func (p *protocol) Read(ctx context.Context) (FastbootStatus, []byte, error) {
	if p.IsClosed {
		return FAIL, nil, FastbootErrors.DeviceClose
	}
	var data []byte
	buf := make([]byte, p.inEndpoint.Desc.MaxPacketSize)
	n, err := p.inEndpoint.ReadContext(ctx, buf)
	if err != nil {
		return FAIL, []byte{}, err
	}
	data = append(data, buf[:n]...)
	return status(data[:4]), data[4:], nil
}

func (p *protocol) Close() {
	if !p.IsClosed {
		p.IsClosed = true
		p.cleanup()
	}
}
