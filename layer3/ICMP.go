package layer3

type ICMPPacket struct {
	msgType     byte
	msgCode     byte
	msgCheckSum uint16
	msgRest     []byte
}

func ProcessICMP(header *Ipv4Header, packet []byte) {

}
