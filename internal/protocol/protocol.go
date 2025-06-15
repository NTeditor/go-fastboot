package protocol

import (
	"context"
	"fmt"

	"github.com/google/gousb"
	"github.com/nteditor/go-fastboot/fastbootErrors"
)

type Protocol struct {
	inEndpoint  *gousb.InEndpoint
	outEndpoint *gousb.OutEndpoint
	cleanup     func()
	IsClosed    bool
}

func NewProtocol(
	inEndpoint *gousb.InEndpoint,
	outEndpoint *gousb.OutEndpoint,
	cleanup func(),
) *Protocol {
	return &Protocol{
		inEndpoint:  inEndpoint,
		outEndpoint: outEndpoint,
		cleanup:     cleanup,
	}
}

func (p *Protocol) Send(ctx context.Context, data []byte) error {
	if p.IsClosed {
		return fastbootErrors.DeviceClose
	}
	_, err := p.outEndpoint.WriteContext(ctx, data)
	return err
}

func (p *Protocol) Read(ctx context.Context) (StatusType, []byte, error) {
	if p.IsClosed {
		return Status.FAIL, nil, fastbootErrors.DeviceClose
	}
	var data []byte
	buf := make([]byte, p.inEndpoint.Desc.MaxPacketSize)
	n, err := p.inEndpoint.ReadContext(ctx, buf)
	if err != nil {
		return Status.FAIL, nil, err
	}
	data = append(data, buf[:n]...)
	return RawToStatus(data[:4]), data[4:], nil
}

func (p *Protocol) Close() {
	if !p.IsClosed {
		p.IsClosed = true
		p.cleanup()
	}
}

func (p *Protocol) Download(ctx context.Context, data []byte) error {
	if p.IsClosed {
		return fastbootErrors.DeviceClose
	}

	const chunkSize = 0x40040
	dataSize := len(data)

	err := p.Send(ctx, []byte(fmt.Sprintf("download:%08x", dataSize)))
	if err != nil {
		return err
	}

	status, _, err := p.Read(ctx)
	switch {
	case status != Status.DATA:
		return fmt.Errorf("failed to start data phase: %s", status)
	case err != nil:
		return err
	}

	for i := 0; i < dataSize; i += chunkSize {
		end := min(i+chunkSize, dataSize)
		err := p.Send(ctx, data[i:end])
		if err != nil {
			return err
		}
	}
	status, resultData, err := p.Read(ctx)
	switch {
	case status != Status.OKAY:
		return fmt.Errorf("failed to finish data phase, status: %s, data: %s", status, resultData)
	case err != nil:
		return err
	}
	return nil
}
