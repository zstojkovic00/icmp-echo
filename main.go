package main

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
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
			fmt.Printf("IP address: %s, mask: %d %d", ipNet.IP.To4(), ones, 32)

			// prvo moramo da izracunamo sa submask sta treba da pingujemo
			startIp, endIp := findNetworkRange(string(ipV4), ones)
			refreshARPTable(startIp, endIp)
		}
	}
}

func refreshARPTable(starIp []int, endIp []int) {

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

	// one ip address return
	if reservedBytes == 4 {
		return startIp, endIp
	}

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
