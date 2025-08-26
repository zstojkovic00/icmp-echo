package main

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	findIpByMac()
}

func findIpByMac() {
	wirelessInterface := findWirelessInterface()
	fmt.Printf("Wireless interface name: " + wirelessInterface.Name)

	addrs, err := wirelessInterface.Addrs()

	if err != nil {
		panic(err)
	}

	// 192.168.1.13/24 pronasli smo ip adresu naseg mreznog interfejsa
	fmt.Printf("\n%+v\n", addrs)

	for _, addr := range addrs {
		ipNet := addr.(*net.IPNet)
		ipV4 := ipNet.IP.To4()
		if ipV4 != nil {
			ones, _ := ipNet.Mask.Size()
			fmt.Printf("IP address: %s, ones: %d bits:%d", ipNet.IP.To4(), ones, 32)

			// prvo moramo da izracunamo sa submask sta treba da pingujemo
			startIp, endIp := findNetworkRange(ipV4.String(), ones)
			fmt.Printf("\n%+v\n", startIp)
			fmt.Printf("%+v\n", endIp)

			// refreshARPTable(startIp, endIp)
		}
	}
}

func refreshARPTable(starIp []int, endIp []int) {
	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)

	if err != nil {
		panic(err)
	}

	icmp := ICMP{
		Type:           8,
		Code:           0,
		Checksum:       0,
		Identifier:     0,
		SequenceNumber: 0,
	}

	packet := icmp.ToBytes()
	checksum := calculateChecksum(packet)
	packet[2] = byte(checksum >> 8)
	packet[3] = byte(checksum)

}

/*
	Packet: [8, 0, 0, 0, 0, 0, 0, 0]

	i=0: pair = 8*256 + 0 = 2048,    checksum = 2048
	i=2: pair = 0*256 + 0 = 0,       checksum = 2048
	i=4: pair = 0*256 + 0 = 0,       checksum = 2048
	i=6: pair = 0*256 + 0 = 0,       checksum = 2048
*/

func calculateChecksum(packet []byte) uint16 {
	var checksum uint32

	for i := 0; i < len(packet); i += 2 {
		pair := uint32(packet[i])*256 + uint32(packet[i+1])
		checksum += pair
	}

	return uint16(^checksum)
}

func findNetworkRange(ipV4 string, ones int) ([]int, []int) {
	// 192.168.1.13/24 => ovo znaci da su prva 24 BITA rezervisana
	// (3 bajta + 0 bita iz 4. bajta), zadnji bajt ide od 0 do 255
	// 192.168.1.0 -> 192.168.1.255

	// 192.168.1.13/25 => ovo znaci da su prva 25 BITA rezervisana
	// (3 bajta + 1 bit iz 4. bajta), zadnjih 7 bita ide od 0 do 127
	// 192.168.1.0 -> 192.168.1.127

	if ones < 24 {
		panic("Too many address combination to ping")
	}

	reservedBytes := ones / 8
	reservedBits := ones % 8
	ipParts := strings.Split(ipV4, ".")

	startIp := make([]int, 4)
	endIp := make([]int, 4)

	for i := 0; i < 4; i++ {
		startIp[i], _ = strconv.Atoi(ipParts[i])
		endIp[i], _ = strconv.Atoi(ipParts[i])
	}

	// jedna ip adresa
	if reservedBytes == 4 {
		return startIp, endIp
	}

	startIp[3] = 0

	if reservedBits == 0 {
		endIp[3] = 255
	} else {
		// (2^8-1)-1 = 127
		availableHosts := int(math.Pow(2, float64(8-reservedBits))) - 1
		endIp[3] = availableHosts
	}

	return startIp, endIp
}

func findWirelessInterface() net.Interface {
	interfaces, err := net.Interfaces()

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", interfaces)

	for _, iface := range interfaces {
		if strings.HasPrefix(iface.Name, "wl") && iface.Flags&net.FlagUp != 0 {
			return iface
		}
	}

	panic("no interface found")
}

type ICMP struct {
	Type           uint8
	Code           uint8
	Checksum       uint16
	Identifier     uint16
	SequenceNumber uint16
}

func (icmp *ICMP) ToBytes() []byte {
	packet := make([]byte, 8)

	packet[0] = icmp.Type
	packet[1] = 0
	packet[2] = 0
	packet[3] = 0
	packet[4] = 0
	packet[5] = 0
	packet[6] = 0
	packet[7] = 0

	return packet
}
