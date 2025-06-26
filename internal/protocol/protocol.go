package protocol

import (
	"context"
	"fmt"
	"io"

	"github.com/google/gousb"
	"github.com/nteditor/go-fastboot/fastbooterrors"
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
		return fastbooterrors.ErrDeviceClose
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
		return fastbooterrors.ErrTimeout
	}
}

func (p *Protocol) Read(ctx context.Context) (StatusType, []byte, error) {
	if p.IsClosed {
		return Status.FAIL, nil, fastbooterrors.ErrDeviceClose
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
		return Status.FAIL, nil, fastbooterrors.ErrTimeout
	}
}

func (p *Protocol) Cleanup() {
	p.cleanup()
}

func (p *Protocol) DownloadReader(ctx context.Context, reader io.Reader, size int64) error {
	if p.IsClosed {
		return fastbooterrors.ErrDeviceClose
	}

	err := p.Send(ctx, []byte(fmt.Sprintf("download:%08x", size)))
	if err != nil {
		return err
	}

	status, statusData, err := p.Read(ctx)
	switch {
	case err != nil:
		return err
	case status == Status.FAIL:
		return &fastbooterrors.ErrStatusFail{Data: statusData}
	}

	const chunk_size = 0x40040
	buf := make([]byte, 0x40040)
	for {
		chunk, err := reader.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		err = p.Send(ctx, buf[:chunk])
		if err != nil {
			return err
		}
	}

	status, statusData, err = p.Read(ctx)
	switch {
	case err != nil:
		return err
	case status == Status.FAIL:
		return &fastbooterrors.ErrStatusFail{Data: statusData}
	}
	return nil
}

func (p *Protocol) Download(ctx context.Context, image []byte) error {
	if p.IsClosed {
		return fastbooterrors.ErrDeviceClose
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
		return &fastbooterrors.ErrStatusFail{Data: statusData}
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
		return &fastbooterrors.ErrStatusFail{Data: statusData}
	}
	return nil
}
