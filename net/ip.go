package net

import "net/netip"

// ParseAddrPtr parses an IP address string and returns a pointer to netip.Addr.
// Returns nil if the string is empty or invalid.
func ParseAddrPtr(s string) *netip.Addr {
	if s == "" {
		return nil
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return nil
	}
	return &addr
}
