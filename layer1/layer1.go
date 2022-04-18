package layer1

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"go_network/utils"
	"net"
	"os"
	"time"
)

const (
	snapLen     = 1600
	promiscuous = true
	timeout     = 30 * time.Second
)

func init() {
	OpenNetDevice(utils.DefaultDev())
}

func OpenNetDevice(device *utils.NetDevice) {
	devices, _ := pcap.FindAllDevs()

	firstDev := devices[0]
	device.Name = firstDev.Name
	copy(device.Ipv4Addr[:], firstDev.Addresses[0].IP)
	copy(device.NetMask[:], firstDev.Addresses[0].Netmask)

	ifs, _ := net.Interfaces()
	for _, ifa := range ifs {
		if ifa.Name == device.Name {
			copy(device.MacAddr[:], ifa.HardwareAddr)
			device.MTU = ifa.MTU
			break
		}
	}

	var err error
	device.Handle, err = pcap.OpenLive(device.Name, snapLen, promiscuous, timeout)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func CloseNetDevice(device *utils.NetDevice) {
	device.Handle.Close()
}

func ReceiveBytes(device *utils.NetDevice) (data []byte) {
	data, _, _ = device.Handle.ReadPacketData()
	return data
}

func SendBytes(device *utils.NetDevice, data []byte) {
	_ = device.Handle.WritePacketData(data)
}
