package main

import (
	"fmt"
	"math/rand"
	"net"
)

func MakeRandomAddress(from, to net.IP) net.IP {
	var result net.IP

	randVal := make([]byte, len(from))
	rand.Read(randVal)

	for i := range from {
		if to[i] == from[i] {
			result = append(result, from[i])
		} else {
			result = append(result, randVal[i]%(to[i]-from[i])+from[i])
		}
	}

	return result
}

func GetAddressByInterface(name string) (*net.IPNet, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		switch x := addr.(type) {
		case *net.IPNet:
			return x, nil
		case *net.IPAddr:
			return &net.IPNet{
				IP:   x.IP,
				Mask: x.IP.DefaultMask(),
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to get address information by %s", name)
}
