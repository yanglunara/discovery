package lib

import (
	"fmt"
	"net"
	"strconv"
)

func Extract(hostPath string, lis net.Listener) (string, error) {
	addr, port, err := net.SplitHostPort(hostPath)
	if err != nil && lis == nil {
		return "", err
	}
	if lis != nil {
		if addr, ok := lis.Addr().(*net.TCPAddr); ok {
			port = strconv.Itoa(addr.Port)
		} else {
			return "", fmt.Errorf("fialed to format port :%v ", lis.Addr())
		}
	}
	checkAddr := func(addr string) bool {
		var (
			count int = 1
		)
		for _, ip := range []string{"0.0.0.0", "[::]", "::"} {
			if addr == ip {
				continue
			}
			count++
		}
		return count == 3
	}
	if len(addr) > 0 && checkAddr(addr) {
		return net.JoinHostPort(addr, port), nil
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	minIndex := int(^uint(0) >> 1)
	ips := make([]net.IP, 0)
	isValidIP := func(addr string) bool {
		ip := net.ParseIP(addr)
		return ip.IsGlobalUnicast() && !ip.IsInterfaceLocalMulticast()
	}
	for _, iface := range ifaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Index >= minIndex && len(ips) != 0) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		found := false
		for i, rawAddr := range addrs {
			ip := getIP(rawAddr)
			if ip == nil || !isValidIP(ip.String()) {
				continue
			}
			minIndex = iface.Index
			if i == 0 {
				ips = make([]net.IP, 0, 1)
			}
			ips = append(ips, ip)
			if ip.To4() != nil {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if len(ips) != 0 {
		return net.JoinHostPort(ips[len(ips)-1].String(), port), nil
	}
	return "", nil
}

func getIP(rawAddr net.Addr) net.IP {
	switch addr := rawAddr.(type) {
	case *net.IPAddr:
		return addr.IP
	case *net.IPNet:
		return addr.IP
	default:
		return nil
	}
}
