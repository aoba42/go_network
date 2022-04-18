package layer4

import (
	"go_network/layer3"
	"go_network/socket"
	"go_network/utils"
)

type UDPHeader struct {
	SourcePort uint16
	DestPort   uint16
	Length     uint16
	CheckSum   uint16
}

type UDPData struct {
	Data []byte
}

func ParseUDPPacket(packet []byte) (header UDPHeader, payload []byte) {
	header.SourcePort = utils.BytesToUint16(packet[0:2])
	header.DestPort = utils.BytesToUint16(packet[2:4])
	header.Length = utils.BytesToUint16(packet[4:6])
	header.CheckSum = utils.BytesToUint16(packet[6:8])
	return header, packet[8:]
}

func ConstructUDPPacket(header UDPHeader, payload []byte) (result []byte) {
	result = append(result, utils.Uint16ToBytes(header.SourcePort)...)
	result = append(result, utils.Uint16ToBytes(header.DestPort)...)
	result = append(result, utils.Uint16ToBytes(header.Length)...)
	result = append(result, utils.Uint16ToBytes(header.CheckSum)...)
	header.CheckSum = utils.CheckSum(result, 8)
	result = append(result, utils.Uint16ToBytes(header.CheckSum)...)
	result = append(result, payload...)
	return
}

func ProcessUDP(header *layer3.Ipv4Header, packet []byte) {
	if utils.CheckSum(packet, 8) != 0xffff {
		return
	}
	udpHeader, payload := ParseUDPPacket(packet)
	udpSocket := socket.UDPSocket{
		DestIp:   header.DestIpAddr,
		DestPort: udpHeader.DestPort,
		SrcIp:    header.SourceIpAddr,
		SrcPort:  udpHeader.SourcePort,
	}
	_, found := socket.UDPSocketRecvChannels[udpSocket]
	if !found {
		socket.UDPSocketRecvChannels[udpSocket] = make(chan UDPData)
	}
	socket.UDPSocketSendChannels[udpSocket] <- UDPData{Data: payload}
}
