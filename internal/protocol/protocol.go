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
		return fastbootErrors.ErrDeviceClose
	}

	resultChan := make(chan error, 1)

	go func() {
		_, err := p.outEndpoint.WriteContext(ctx, data)
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return fastbootErrors.ErrTimeout
	}
}

func (p *Protocol) Read(ctx context.Context) (StatusType, []byte, error) {
	if p.IsClosed {
		return Status.FAIL, nil, fastbootErrors.ErrDeviceClose
	}

	type resultStruct struct {
		status StatusType
		data   []byte
		err    error
	}

	resultChan := make(chan resultStruct, 1)

	go func() {
		var data []byte
		buf := make([]byte, p.inEndpoint.Desc.MaxPacketSize)
		n, err := p.inEndpoint.ReadContext(ctx, buf)
		if err != nil {
			resultChan <- resultStruct{status: Status.FAIL, data: nil, err: err}
		}
		data = append(data, buf[:n]...)
		resultChan <- resultStruct{status: RawToStatus(data[:4]), data: data[4:], err: nil}
	}()

	select {
	case result := <-resultChan:
		return result.status, result.data, result.err
	case <-ctx.Done():
		return Status.FAIL, nil, fastbootErrors.ErrTimeout
	}
}

func (p *Protocol) Cleanup() {
	p.cleanup()
}

func (p *Protocol) Download(ctx context.Context, image []byte) error {
	if p.IsClosed {
		return fastbootErrors.ErrDeviceClose
	}

	const chunk_size = 0x40040
	dataSize := len(image)

	err := p.Send(ctx, []byte(fmt.Sprintf("download:%08x", dataSize)))
	if err != nil {
		return err
	}

	status, statusData, err := p.Read(ctx)
	switch {
	case err != nil:
		return err
	case status == Status.FAIL:
		return &fastbootErrors.ErrStatusFail{Data: statusData}
	}

	for i := 0; i < dataSize; i += chunk_size {
		end := min(i+chunk_size, dataSize)
		err := p.Send(ctx, image[i:end])
		if err != nil {
			return err
		}
	}
	status, statusData, err = p.Read(ctx)
	switch {
	case err != nil:
		return err
	case status == Status.FAIL:
		return &fastbootErrors.ErrStatusFail{Data: statusData}
	}
	return nil
}
