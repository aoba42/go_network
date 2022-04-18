package socket

import (
	"go_network/layer4"
	"go_network/utils"
)

type UDPSocket struct {
	DestIp   [4]byte
	DestPort uint16
	SrcIp    [4]byte
	SrcPort  uint16
}

var UDPSocketRecvChannels map[UDPSocket]chan layer4.UDPData
var UDPSocketSendChannels map[UDPSocket]chan layer4.UDPData

func UDPSocketRecv(destIp [4]byte, destPort uint16, srcPort uint16) []byte {
	socket := UDPSocket{
		DestIp:   destIp,
		DestPort: destPort,
		SrcIp:    utils.DefaultDev().Ipv4Addr,
		SrcPort:  srcPort,
	}
	_, found := UDPSocketRecvChannels[socket]
	if !found {
		UDPSocketRecvChannels[socket] = make(chan layer4.UDPData)
	}
	payload := <-UDPSocketRecvChannels[socket]
	return payload.Data
}

func UDPSocketSend(destIp [4]byte, destPort uint16, srcPort uint16, data []byte) {
	socket := UDPSocket{
		DestIp:   destIp,
		DestPort: destPort,
		SrcIp:    utils.DefaultDev().Ipv4Addr,
		SrcPort:  srcPort,
	}
	_, found := UDPSocketSendChannels[socket]
	if !found {
		UDPSocketSendChannels[socket] = make(chan layer4.UDPData)
	}
	UDPSocketSendChannels[socket] <- layer4.UDPData{Data: data}
}
