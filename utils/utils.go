package utils

import "github.com/google/gopacket/pcap"

type NetDevice struct {
	Name     string
	Ipv4Addr [4]byte
	NetMask  [4]byte
	MTU      int
	MacAddr  [6]byte
	Handle   *pcap.Handle
	GateWay  [4]byte
}

var BoardCastMac = [6]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

var defaultDev NetDevice

func DefaultDev() *NetDevice {
	return &defaultDev
}

func CheckSum(data []byte, count int) uint16 {
	var sum uint32 = 0
	for i := 0; i*2 < count; i++ {
		sum += uint32(data[i*2])<<8 + uint32(data[i*2+1])
	}
	return uint16(sum & 0xffff)
}

func BytesToUint16(data []byte) uint16 {
	return uint16(data[0])<<8 + uint16(data[1])
}

func BytesToUint32(data []byte) uint32 {
	return uint32(data[0])<<8 + uint32(data[1])
}

func Uint16ToBytes(data uint16) []byte {
	res := [2]byte{
		byte((data >> 8) & 0xff),
		byte(data & 0xff),
	}
	return res[:]
}

func Uint32ToBytes(data uint32) []byte {
	res := [4]byte{
		byte((data >> 24) & 0xff),
		byte((data >> 16) & 0xff),
		byte((data >> 8) & 0xff),
		byte(data & 0xff),
	}
	return res[:]
}
