// Package proto provides the means to parse and generate frames of the KNXnet/IP protocol.
package proto

import (
	"bytes"
	"errors"
	"io"

	"github.com/vapourismo/knx-go/knx/encoding"
	"github.com/vapourismo/knx-go/knx/util"
)

// ServiceID identifies the service that is contained in a packet.
type ServiceID uint16

// These are supported services.
const (
	ConnReqService      ServiceID = 0x0205
	ConnResService      ServiceID = 0x0206
	ConnStateReqService ServiceID = 0x0207
	ConnStateResService ServiceID = 0x0208
	DiscReqService      ServiceID = 0x0209
	DiscResService      ServiceID = 0x020a
	TunnelReqService    ServiceID = 0x0420
	TunnelResService    ServiceID = 0x0421
	RoutingIndService   ServiceID = 0x0530
	RoutingLostService  ServiceID = 0x0531
	RoutingBusyService  ServiceID = 0x0532
)

// Service describes a KNXnet/IP service.
type Service interface {
	Service() ServiceID
}

// ServiceWriterTo combines WriterTo and Service.
type ServiceWriterTo interface {
	Service
	io.WriterTo
}

// Pack generates a KNXnet/IP packet.
func Pack(w io.Writer, srv ServiceWriterTo) (int64, error) {
	dataBuffer := bytes.Buffer{}

	_, err := srv.WriteTo(&dataBuffer)
	if err != nil {
		return 0, err
	}

	return encoding.WriteSome(
		w, byte(6), byte(16), srv.Service(), uint16(dataBuffer.Len()+6), &dataBuffer,
	)
}

// These are errors that might occur during unpacking.
var (
	ErrHeaderLength   = errors.New("Header length is not 6")
	ErrHeaderVersion  = errors.New("Protocol version is not 16")
	ErrUnknownService = errors.New("Unknown service identifier")
)

type serviceUnpackable interface {
	util.Unpackable
	Service
}

// Unpack parses a KNXnet/IP packet and retrieves its service payload.
//
// On success, the variable pointed to by srv will contain a pointer to a service type.
// You can cast it to the matching against service type, like so:
//
// 	var srv Service
//
// 	_, err := Unpack(r, &srv)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	switch srv := srv.(type) {
// 		case *ConnRes:
// 			// ...
//
// 		case *TunnelReq:
// 			// ...
//
// 		// ...
// 	}
//
func Unpack(data []byte, srv *Service) (uint, error) {
	var headerLen, version uint8
	var srvID ServiceID
	var totalLen uint16

	n, err := util.UnpackSome(data, &headerLen, &version, &srvID, &totalLen)
	if err != nil {
		return n, err
	}

	if headerLen != 6 {
		return n, ErrHeaderLength
	}

	if version != 16 {
		return n, ErrHeaderVersion
	}

	var body serviceUnpackable
	switch srvID {
	case ConnReqService:
		body = &ConnReq{}

	case ConnResService:
		body = &ConnRes{}

	case ConnStateReqService:
		body = &ConnStateReq{}

	case ConnStateResService:
		body = &ConnStateRes{}

	case DiscReqService:
		body = &DiscReq{}

	case DiscResService:
		body = &DiscRes{}

	case TunnelReqService:
		body = &TunnelReq{}

	case TunnelResService:
		body = &TunnelRes{}

	case RoutingIndService:
		body = &RoutingInd{}

	case RoutingLostService:
		body = &RoutingLost{}

	case RoutingBusyService:
		body = &RoutingBusy{}

	default:
		return n, ErrUnknownService
	}

	m, err := body.Unpack(data[n:])

	if err == nil {
		*srv = body
	}

	return n + m, err
}