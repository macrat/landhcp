package main

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type LeaseInfo struct {
	Address net.IP
	Client  net.HardwareAddr
	Name    string
	Expire  time.Time
}

type AddressManager interface {
	SubnetMask() net.IPMask
	LeaseTimeFor(net.HardwareAddr) time.Duration
	MakeOffer(hw net.HardwareAddr, name string) (net.IP, error)
	Acquire(ip net.IP, hw net.HardwareAddr, name string) error
	Release(ip net.IP, hw net.HardwareAddr, name string) error
}

type InMemoryAddressManager struct {
	rangeFrom  net.IP
	rangeTo    net.IP
	subnetMask net.IPMask
	leaseTime  time.Duration
	leases     map[string]LeaseInfo
}

func NewInMemoryAddressManager(from, to net.IP, mask net.IPMask, leaseTime time.Duration) *InMemoryAddressManager {
	return &InMemoryAddressManager{
		rangeFrom:  from,
		rangeTo:    to,
		subnetMask: mask,
		leaseTime:  leaseTime,
		leases:     make(map[string]LeaseInfo),
	}
}

func (m *InMemoryAddressManager) SubnetMask() net.IPMask {
	return m.subnetMask
}

func (m *InMemoryAddressManager) LeaseTimeFor(client net.HardwareAddr) time.Duration {
	for _, r := range m.leases {
		if bytes.Equal(r.Client, client) {
			t := r.Expire.Sub(time.Now())
			if t > 0 {
				return t
			}
		}
	}
	return m.leaseTime
}

func (m *InMemoryAddressManager) MakeOffer(hw net.HardwareAddr, name string) (net.IP, error) {
	for _, r := range m.leases {
		if bytes.Equal(r.Client, hw) {
			return r.Address, nil
		}
	}

	for {
		addr := MakeRandomAddress(m.rangeFrom, m.rangeTo)

		info, exists := m.leases[addr.String()]
		if exists && info.Expire.After(time.Now()) {
			continue
		}

		return addr, nil
	}
}

func (m *InMemoryAddressManager) Acquire(ip net.IP, hw net.HardwareAddr, name string) error {
	for _, r := range m.leases {
		if r.Address.Equal(ip) {
			if bytes.Equal(r.Client, hw) {
				break
			}
			return fmt.Errorf("%s is already acquired by another host", ip)
		}
	}
	m.leases[ip.String()] = LeaseInfo{
		Address: ip,
		Client:  hw,
		Name:    name,
		Expire:  time.Now().Add(m.leaseTime),
	}
	return nil
}

func (m *InMemoryAddressManager) Release(ip net.IP, hw net.HardwareAddr, name string) error {
	if r, ok := m.leases[ip.String()]; !ok || !bytes.Equal(r.Client, hw) {
		return fmt.Errorf("%s is not acquired by %s", ip, hw)
	}
	delete(m.leases, ip.String())
	return nil
}
