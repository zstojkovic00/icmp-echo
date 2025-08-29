package main

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"syscall"

	"icmp-echo/icmp"
)

func main() {
	startIP, endIP := scanLocalNetwork()
	pingIPRange(startIP, endIP)
}

func scanLocalNetwork() ([]int, []int) {
	wirelessInterface := findWirelessInterface()
	fmt.Printf("Wireless interface name: %s\n", wirelessInterface.Name)

	addrs, err := wirelessInterface.Addrs()
	if err != nil {
		panic(err)
	}

	// Found IP address of our network interface (e.g., 192.168.1.13/24)
	fmt.Printf("Interface addresses: %+v\n", addrs)

	for _, addr := range addrs {
		ipNet := addr.(*net.IPNet)
		ipV4 := ipNet.IP.To4()
		if ipV4 != nil {
			ones, _ := ipNet.Mask.Size()
			fmt.Printf("IP address: %s, subnet bits: %d, total bits: %d\n", ipV4, ones, 32)

			// Calculate subnet range to determine what to ping
			startIP, endIP := calculateIPRange(ipV4.String(), ones)
			fmt.Printf("Network range start: %+v\n", startIP)
			fmt.Printf("Network range end: %+v\n", endIP)

			return startIP, endIP
		}
	}

	panic("No IPv4 address found on wireless interface")
}

func pingIPRange(startIP []int, endIP []int) {
	socket, err := syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_RAW,
		syscall.IPPROTO_ICMP,
	)
	defer syscall.Close(socket)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Socket creation succeeded\n")

	// Create ICMP Echo Request packet
	icmpMsg := icmp.NewEchoRequest(0, nil)
	packet := icmpMsg.Serialize()

	for i := startIP[3]; i <= endIP[3]; i++ {
		addr := &syscall.SockaddrInet4{
			Port: 0,
			Addr: [4]byte{
				byte(startIP[0]),
				byte(startIP[1]),
				byte(startIP[2]),
				byte(i),
			},
		}

		fmt.Printf("Pinging %d.%d.%d.%d\n", startIP[0], startIP[1], startIP[2], i)

		err = syscall.Sendto(socket, packet, 0, addr)
		if err != nil {
			fmt.Printf("Error while sending ping packet: %v\n", err)
			return
		}
		fmt.Printf("ICMP packet sent successfully\n")

		// buffer := make([]byte, 1500)
		// n, _, err := syscall.Recvfrom(socket, buffer, 0)
		// if err != nil {
		//     fmt.Printf("Error while receiving ping packet: %v\n", err)
		// } else {
		//     fmt.Printf("ICMP packet received: %d bytes\n", n)
		// }
	}

	fmt.Printf("Network scan completed")
}

/*
	192.168.1.13/24 => means first 24 BITS are reserved for network
	(3 bytes + 0 bits from 4th byte), last byte ranges from 0 to 255
	192.168.1.0 -> 192.168.1.255

	192.168.1.13/25 => means first 25 BITS are reserved for network
	(3 bytes + 1 bit from 4th byte), last 7 bits range from 0 to 127
	192.168.1.0 -> 192.168.1.127
*/

func calculateIPRange(ipV4 string, ones int) ([]int, []int) {

	if ones < 24 {
		panic("Subnet too large, networks smaller than /24 would require pinging too many hosts")
	}

	if ones > 32 {
		panic("Invalid subnet prefix, cannot exceed 32 bits for IPv4")
	}

	reservedBytes := ones / 8
	reservedBits := ones % 8
	ipParts := strings.Split(ipV4, ".")

	startIP := make([]int, 4)
	endIP := make([]int, 4)

	for i := 0; i < 4; i++ {
		startIP[i], _ = strconv.Atoi(ipParts[i])
		endIP[i], _ = strconv.Atoi(ipParts[i])
	}

	// Single IP address case
	if reservedBytes == 4 {
		return startIP, endIP
	}

	startIP[3] = 1 // Start from .1 (skip network address .0)

	if reservedBits == 0 {
		endIP[3] = 254 // End at .254 (skip broadcast address .255)
	} else {
		// Calculate available host addresses: (2^host_bits) - 2
		availableHosts := int(math.Pow(2, float64(8-reservedBits))) - 2
		endIP[3] = availableHosts
	}

	return startIP, endIP
}

func findWirelessInterface() net.Interface {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Available interfaces: %+v\n", interfaces)

	for _, iface := range interfaces {
		if strings.HasPrefix(iface.Name, "wl") &&
			iface.Flags&net.FlagUp != 0 {
			return iface
		}
	}

	panic("No wireless interface found")
}
