package layer3

import (
	"go_network/layer2"
	"go_network/layer4"
	"go_network/utils"
)

type Ipv4Header struct {
	Version        byte
	headerLen      byte
	ServiceType    byte
	TotalLength    uint16
	Identifier     uint16
	FragmentFlags  byte
	FragmentOffset uint16
	TimeToLive     byte
	Protocol       byte
	HeaderCheckSum uint16
	SourceIpAddr   [4]byte
	DestIpAddr     [4]byte
	Options        []byte //exist only header length > 20 bytes
}

const (
	ProtocolICMP = 1
	ProtocolTCP  = 6
	ProtocolUDP  = 17
	fragFlagMask = 0x03
	fragFlagDont = 0x02
	fragFlagMore = 0x01
)

func ParseIpv4PacketHeader(packet []byte) (header Ipv4Header, payload []byte) {
	header.Version = (packet[0] >> 4) & 0x0F
	header.headerLen = packet[0] & 0x0F
	header.ServiceType = packet[1]
	header.TotalLength = utils.BytesToUint16(packet[2:4])
	header.Identifier = utils.BytesToUint16(packet[4:6])
	header.FragmentFlags = packet[6] & 0b11100000
	offsetBytes := []byte{
		packet[6] & 0b00011111,
		packet[7],
	}
	header.FragmentOffset = utils.BytesToUint16(offsetBytes)
	header.TimeToLive = packet[8]
	header.Protocol = packet[9]
	header.HeaderCheckSum = utils.BytesToUint16(packet[10:12])
	copy(header.DestIpAddr[:], packet[12:16])
	copy(header.SourceIpAddr[:], packet[16:20])
	if header.headerLen*4 > 20 {
		header.Options = append(header.Options, packet[20:header.headerLen*4]...)
	}
	payload = packet[header.headerLen*4:]
	return
}

func ConstructIpv4Packet(header Ipv4Header, payload []byte) (result []byte) {
	result = append(result, header.Version<<4|header.headerLen)
	result = append(result, header.ServiceType)
	result = append(result, utils.Uint16ToBytes(header.TotalLength)...)
	result = append(result, utils.Uint16ToBytes(header.Identifier)...)
	result = append(result, header.FragmentFlags<<5|byte(header.FragmentOffset>>8))
	result = append(result, byte(header.FragmentOffset&0x00ff))
	result = append(result, header.TimeToLive)
	result = append(result, header.Protocol)
	result = append(result, utils.Uint16ToBytes(header.HeaderCheckSum)...)
	result = append(result, header.DestIpAddr[:]...)
	result = append(result, header.SourceIpAddr[:]...)
	result = append(result, header.Options[:]...)
	result = append(result, payload...)
	return
}

/*
type Fragment struct {
	start int
	data  []byte
	next  *Fragment
}

type FragmentHeader struct {
	header       Ipv4Header
	lastArrived  bool
	fragmentList *Fragment
}

var fragmentBuffer map[uint16]FragmentHeader

func init() {
	fragmentBuffer = make(map[uint16]FragmentHeader)
}
*/
func ProcessIpv4(device *utils.NetDevice, packet []byte) {
	header, payload := ParseIpv4PacketHeader(packet)
	if utils.CheckSum(packet, int(header.headerLen)*4) != 0xffff {
		return
	}
	if header.DestIpAddr != device.Ipv4Addr {
		if header.TimeToLive == 0 {
			return
		}
		header.TimeToLive -= 1
		nextIpAddr := searchRoutingTable(device, header.DestIpAddr)
		nextMacAddr, found := layer2.SearchArpTable(nextIpAddr)
		if !found {
			nextMacAddr = layer2.ArpQueryUnknownIp(device, nextIpAddr)
		}
		if device.MTU >= int(header.TotalLength) {
			layer2.UniCastSend(device, packet, nextMacAddr, layer2.EtherTypeIp)
		}
		/*else {
			fragments := fragile(device, &header, payload)
			for _, frag := range fragments {
				layer2.UniCastSend(device, frag, nextMacAddr, layer2.EtherTypeIp)
			}
		}*/
	} else {
		/*fragHeader, found := fragmentBuffer[header.Identifier]
		if found || header.FragmentFlags&fragFlagMask == fragFlagMore {
			mergeFragment(&fragHeader, &header, payload)
			if header.FragmentFlags&fragFlagMask == fragFlagDont &&
				fragHeader.fragmentList != nil &&
				fragHeader.fragmentList.next == nil {
				DispatchIpv4(&header, payload)
			}
		} else {
			DispatchIpv4(&header, payload)
		}*/
		if header.FragmentFlags&fragFlagMask == fragFlagDont {
			DispatchIpv4(&header, payload)
		}
	}
}

func searchRoutingTable(device *utils.NetDevice, ip [4]byte) [4]byte {
	// simplify the routing protocol
	return device.GateWay
}

/*
func fragile(device *utils.NetDevice, header *Ipv4Header, payload []byte) [][]byte {
	const etherHeaderLen = 14
	maxPayloadLen := device.MTU - int(header.headerLen)*4 - etherHeaderLen
	maxPayloadLen = maxPayloadLen / 8 * 8
	var result [][]byte
	var end = header.FragmentOffset*8 + uint16(len(payload))
	for start := 0; start < len(payload); start += maxPayloadLen {
		newHeader := *header
		fragLen := 0
		if uint16(start+maxPayloadLen) < end {
			fragLen = maxPayloadLen
			newHeader.FragmentFlags = fragFlagMore
		} else {
			fragLen = int(end) - start
			newHeader.FragmentFlags = fragFlagDont
		}
		newHeader.TotalLength = uint16(header.headerLen) + uint16(fragLen)
		newHeader.FragmentOffset = header.FragmentOffset + uint16(start/8)
		newHeader.HeaderCheckSum = 0
		data := ConstructIpv4Packet(newHeader, []byte{})
		newHeader.HeaderCheckSum = utils.CheckSum(data, int(newHeader.headerLen))
		data = ConstructIpv4Packet(newHeader, payload[start:start+fragLen])
		result = append(result, data)
	}
	return result
}

func mergeFragment(fragHeader *FragmentHeader, packetHeader *Ipv4Header, payload []byte) {
	fragStart := packetHeader.FragmentOffset * 8
	fragEnd := fragStart + uint16(len(payload))
	for ptr := fragHeader.fragmentList; ptr != nil; ptr = ptr.next {
		start, end := uint16(ptr.start), uint16(ptr.start+len(ptr.data))
		if fragStart < start && start < fragEnd {
			ptr.data = append(payload[:start-fragStart], ptr.data...)
		}
	}
}
*/
func DispatchIpv4(header *Ipv4Header, payload []byte) {
	if header.Protocol == ProtocolICMP {
		ProcessICMP(header, payload)
	} else if header.Protocol == ProtocolUDP {
		layer4.ProcessUDP(header, payload)
	} else if header.Protocol == ProtocolTCP {

	}
}
