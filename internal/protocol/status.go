package protocol

type StatusType string

var Status = struct {
	OKAY StatusType
	INFO StatusType
	DATA StatusType
	FAIL StatusType
}{
	OKAY: "OKAY",
	INFO: "INFO",
	DATA: "DATA",
	FAIL: "FAIL",
}

func RawToStatus(raw []byte) StatusType {
	switch string(raw[:4]) {
	case "OKAY":
		return Status.OKAY
	case "INFO":
		return Status.INFO
	case "DATA":
		return Status.DATA
	default:
		return Status.FAIL
	}

}
