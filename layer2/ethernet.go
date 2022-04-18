package layer2

import (
	"go_network/layer1"
	"go_network/layer3"
	"go_network/utils"
)

type EtherNetHeader struct {
	DestMac   [6]byte
	SrcMac    [6]byte
	EtherType uint16
}

const (
	headerSize   = 14
	EtherTypeIp  = 0x0800
	EtherTypeArp = 0x0806
)

func ParseEtherNetHeader(data []byte) (header EtherNetHeader) {
	copy(header.DestMac[:], data[0:6])
	copy(header.SrcMac[:], data[6:12])
	header.EtherType = utils.BytesToUint16(data[12:])
	return header
}

func ConstructEtherNetPacket(payload []byte, header *EtherNetHeader) []byte {
	frame := [14]byte{}
	copy(frame[0:6], header.DestMac[:])
	copy(frame[6:12], header.SrcMac[:])
	copy(frame[12:14], utils.Uint16ToBytes(header.EtherType))
	return append(frame[:], payload...)
}

func Layer2Dispatch(device *utils.NetDevice, packet []byte) {
	header := ParseEtherNetHeader(packet)
	payload := packet[headerSize:]
	if header.DestMac == device.MacAddr ||
		header.DestMac == utils.BoardCastMac {
		if header.EtherType == EtherTypeIp {
			layer3.ProcessIpv4(device, payload)
		} else if header.EtherType == EtherTypeArp {
			ProcessArpIpv4(device, payload)
		}
	}
}

func UniCastSend(device *utils.NetDevice, payload []byte,
	destMac [6]byte, etherType uint16) {
	header := EtherNetHeader{
		DestMac:   destMac,
		SrcMac:    device.MacAddr,
		EtherType: etherType,
	}
	packet := ConstructEtherNetPacket(payload, &header)
	layer1.SendBytes(device, packet)
}

func BroadCastSend(device *utils.NetDevice, payload []byte, etherType uint16) {
	header := EtherNetHeader{
		DestMac:   utils.BoardCastMac,
		SrcMac:    device.MacAddr,
		EtherType: etherType,
	}
	packet := ConstructEtherNetPacket(payload, &header)
	layer1.SendBytes(device, packet)
}
