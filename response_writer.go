package main

import (
	"github.com/insomniacslk/dhcp/dhcpv4"
)

type ResponseWriter interface {
	SetType(dhcpv4.MessageType)
	AddModifiers(...dhcpv4.Modifier)
	AddOptions(...dhcpv4.Option)
}

type V4ResponseWriter struct {
	modifiers []dhcpv4.Modifier
}

func (w *V4ResponseWriter) SetType(t dhcpv4.MessageType) {
	w.AddModifiers(dhcpv4.WithMessageType(t))
}

func (w *V4ResponseWriter) AddModifiers(ms ...dhcpv4.Modifier) {
	w.modifiers = append(w.modifiers, ms...)
}

func (w *V4ResponseWriter) AddOptions(os ...dhcpv4.Option) {
	for _, o := range os {
		w.modifiers = append(w.modifiers, dhcpv4.WithOption(o))
	}
}
