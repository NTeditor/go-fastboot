package fastboot

import (
	"github.com/google/gousb"
)

type FastbootStatus string

var Status = struct {
	OKAY FastbootStatus
	FAIL FastbootStatus
	DATA FastbootStatus
	INFO FastbootStatus
}{
	OKAY: "OKAY",
	FAIL: "FAIL",
	DATA: "DATA",
	INFO: "INFO",
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

func (p *protocol) Send(data []byte) error {
	_, err := p.outEndpoint.Write(data)
	return err
}

func (p *protocol) Read() (string, []byte, error) {
	var data []byte
	buf := make([]byte, p.inEndpoint.Desc.MaxPacketSize)
	n, err := p.inEndpoint.Read(buf)
	if err != nil {
		return "FAIL", []byte{}, err
	}
	data = append(data, buf[:n]...)
	status := string(data[:4])
	return status, data[4:], nil
}

func (p *protocol) Close() {
	if !p.IsClosed {
		p.IsClosed = true
		p.cleanup()
		p.inEndpoint = nil
		p.outEndpoint = nil
	}
}
