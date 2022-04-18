package layer2

import (
	"go_network/utils"
	"sync"
)

type ArpIpv4Packet struct {
	hardwareType    uint16
	protocolType    uint16
	hardwareAddrLen byte
	protocolAddrLen byte
	operation       uint16
	senderMacAddr   [6]byte
	senderIpAddr    [4]byte
	targetMacAddr   [6]byte
	targetIpAddr    [4]byte
}

func ParseArpIpv4Packet(data []byte) ArpIpv4Packet {
	var arp ArpIpv4Packet
	arp.hardwareType = utils.BytesToUint16(data[0:2])
	arp.protocolType = utils.BytesToUint16(data[2:4])
	arp.hardwareAddrLen = data[4]
	arp.protocolAddrLen = data[5]
	arp.operation = utils.BytesToUint16(data[6:8])
	copy(arp.senderMacAddr[:], data[8:14])
	copy(arp.senderIpAddr[:], data[14:18])
	copy(arp.targetMacAddr[:], data[18:24])
	copy(arp.targetIpAddr[:], data[24:28])
	return arp
}

func ConstructArpIpv4Packet(arp *ArpIpv4Packet) []byte {
	var res []byte
	res = append(res, utils.Uint16ToBytes(arp.hardwareType)...)
	res = append(res, utils.Uint16ToBytes(arp.protocolType)...)
	res = append(res, arp.hardwareAddrLen)
	res = append(res, arp.protocolAddrLen)
	res = append(res, utils.Uint16ToBytes(arp.operation)...)
	res = append(res, arp.senderMacAddr[:]...)
	res = append(res, arp.senderIpAddr[:]...)
	res = append(res, arp.targetMacAddr[:]...)
	res = append(res, arp.targetIpAddr[:]...)
	return res
}

var arpIpv4Table map[[4]byte][6]byte
var arpQueryWait map[[4]byte]sync.WaitGroup

func init() {
	arpIpv4Table = make(map[[4]byte][6]byte)
	arpQueryWait = make(map[[4]byte]sync.WaitGroup)
}

const (
	operationRequest     = 1
	operationReply       = 2
	HardwareTypeEthernet = 1
	ProtocolTypeIpv4     = 0x0800
)

func ProcessArpIpv4(device *utils.NetDevice, packet []byte) {
	header := ParseArpIpv4Packet(packet)

	if header.hardwareType == HardwareTypeEthernet &&
		header.protocolType == ProtocolTypeIpv4 {
		arpIpv4Table[header.senderIpAddr] = header.senderMacAddr
		if device.Ipv4Addr == header.targetIpAddr {
			if header.operation == operationRequest {
				header.targetMacAddr = header.senderMacAddr
				header.targetIpAddr = header.senderIpAddr
				header.senderMacAddr = device.MacAddr
				header.senderIpAddr = device.Ipv4Addr
				header.operation = operationReply
				replyPacket := ConstructArpIpv4Packet(&header)
				UniCastSend(device, replyPacket, header.targetMacAddr, EtherTypeArp)
			} else if header.operation == operationReply {
				arpIpv4Table[header.senderIpAddr] = header.senderMacAddr
				wg, found := arpQueryWait[header.senderIpAddr]
				if found {
					wg.Done()
				}
			}
		}
	}
}

func SearchArpTable(ipAddr [4]byte) (mac [6]byte, found bool) {
	mac, found = arpIpv4Table[ipAddr]
	return
}

func ArpQueryUnknownIp(device *utils.NetDevice, ipAddr [4]byte) [6]byte {
	arp := ArpIpv4Packet{
		hardwareType:    HardwareTypeEthernet,
		protocolType:    ProtocolTypeIpv4,
		hardwareAddrLen: byte(6),
		protocolAddrLen: byte(4),
		operation:       operationRequest,
		senderMacAddr:   device.MacAddr,
		senderIpAddr:    device.Ipv4Addr,
		targetMacAddr:   [6]byte{0, 0, 0, 0, 0, 0},
		targetIpAddr:    ipAddr,
	}
	packet := ConstructArpIpv4Packet(&arp)
	wg, found := arpQueryWait[ipAddr]
	if !found {
		arpQueryWait[ipAddr] = sync.WaitGroup{}
		wg, _ := arpQueryWait[ipAddr]
		wg.Add(1)
	}
	BroadCastSend(device, packet, EtherTypeArp)
	wg, _ = arpQueryWait[ipAddr]
	wg.Wait()
	return arpIpv4Table[ipAddr]
}
