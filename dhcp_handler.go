package main

import (
	"fmt"
	"log"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

type DHCPHandler struct {
	ServerIP       net.IP
	Routers        []net.IP
	NameServers    []net.IP
	AddressManager AddressManager
}

func (h DHCPHandler) OnDiscover(r *dhcpv4.DHCPv4, w ResponseWriter) error {
	w.SetType(dhcpv4.MessageTypeOffer)

	ip, err := h.AddressManager.MakeOffer(r.ClientHWAddr, r.HostName())
	if err != nil {
		return fmt.Errorf("failed to make offer: %s", err)
	}

	w.AddModifiers(dhcpv4.WithYourIP(ip))

	return nil
}

func (h DHCPHandler) OnRequest(r *dhcpv4.DHCPv4, w ResponseWriter) error {
	reqIP := r.RequestedIPAddress()

	if err := h.AddressManager.Acquire(reqIP, r.ClientHWAddr, r.HostName()); err != nil {
		log.Printf("failed to acquire %s: %s", reqIP, err)
		w.SetType(dhcpv4.MessageTypeNak)
	} else {
		log.Printf("acquire %s for %s", reqIP, r.ClientHWAddr)
		w.SetType(dhcpv4.MessageTypeAck)
		w.AddModifiers(dhcpv4.WithYourIP(reqIP))
	}

	return nil // always nil because reply something even if failed to acquire
}

func (h DHCPHandler) OnRelease(r *dhcpv4.DHCPv4, w ResponseWriter) error {
	if err := h.AddressManager.Release(r.ClientIPAddr, r.ClientHWAddr, r.HostName()); err != nil {
		return fmt.Errorf("failed to release %s: %s", r.ClientIPAddr, err)
	} else {
		log.Printf("release %s from %s", r.ClientIPAddr, r.ClientHWAddr)
	}
	return nil
}

func (h DHCPHandler) SetInformations(r *dhcpv4.DHCPv4, w ResponseWriter) {
	for _, op := range r.ParameterRequestList() {
		switch op {
		case dhcpv4.OptionSubnetMask:
			w.AddOptions(dhcpv4.OptSubnetMask(h.AddressManager.SubnetMask()))
		case dhcpv4.OptionRouter:
			w.AddOptions(dhcpv4.OptRouter(h.Routers...))
		case dhcpv4.OptionNameServer:
			w.AddOptions(dhcpv4.OptDNS(h.NameServers...))
		case dhcpv4.OptionIPAddressLeaseTime:
			w.AddOptions(dhcpv4.OptIPAddressLeaseTime(h.AddressManager.LeaseTimeFor(r.ClientHWAddr)))
		case dhcpv4.OptionServerIdentifier:
			w.AddOptions(dhcpv4.OptServerIdentifier(h.ServerIP))
		}
	}
}

func (h DHCPHandler) Handler(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
	resp := &V4ResponseWriter{}
	resp.AddModifiers(dhcpv4.WithServerIP(h.ServerIP))

	switch req.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		if err := h.OnDiscover(req, resp); err != nil {
			log.Print(err)
			return
		}
	case dhcpv4.MessageTypeRequest:
		if err := h.OnRequest(req, resp); err != nil {
			log.Print(err)
			return
		}
	case dhcpv4.MessageTypeRelease:
		if err := h.OnRelease(req, resp); err != nil {
			log.Print(err)
		}
		return // never reply for release rquest
	case dhcpv4.MessageTypeInform:
		resp.SetType(dhcpv4.MessageTypeAck)
	case dhcpv4.MessageTypeDecline:
		return // ignore decline message
	default:
		log.Printf("unsupported incoming message type:%s from %s(%s)", req.MessageType(), peer, req.ClientHWAddr)
	}

	h.SetInformations(req, resp)

	msg, err := dhcpv4.NewReplyFromRequest(req, resp.modifiers...)
	if err != nil {
		log.Printf("failed to build message: %s", err)
		return
	}

	conn.WriteTo(msg.ToBytes(), peer)
}
