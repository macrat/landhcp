package main

import (
	"log"
	"net"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

var (
	iface       = kingpin.Flag("interface", "Interface name.").Short('i').Default("eth0").String()
	rangeFrom   = kingpin.Flag("range-from", "Lease address range from.").Short('f').Required().IP()
	rangeTo     = kingpin.Flag("range-to", "Lease address range to.").Short('t').Required().IP()
	routers     = kingpin.Flag("routers", "Router addresses.").IPList()
	listen      = kingpin.Flag("listen", "Listen address.").Default("0.0.0.0:67").TCP()
	nameServers = kingpin.Flag("name-servers", "Name server addresses.").IPList()
)

func main() {
	kingpin.Parse()

	network, err := GetAddressByInterface(*iface)
	if err != nil {
		log.Fatal(err)
	}

	if len(*routers) == 0 {
		routers = &[]net.IP{network.IP}
	}

	if len(*nameServers) == 0 {
		routers = &[]net.IP{network.IP}
	}

	listenAddr := &net.UDPAddr{
		IP:   (*listen).IP,
		Port: (*listen).Port,
	}
	h := DHCPHandler{
		ServerIP: network.IP,
		AddressManager: NewInMemoryAddressManager(
			*rangeFrom,
			*rangeTo,
			network.Mask,
			10*time.Minute,
		),
		Routers:     *routers,
		NameServers: *nameServers,
	}
	server, err := server4.NewServer("br0", listenAddr, h.Handler)
	if err != nil {
		log.Fatal(err)
	}

	cidr, _ := network.Mask.Size()
	log.Printf("start listen on %s for lease %s/%d-%s/%d", listenAddr, rangeFrom, cidr, rangeTo, cidr)

	server.Serve()
}
