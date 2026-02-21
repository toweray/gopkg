package iploc

import (
	"net/netip"

	"github.com/phuslu/iploc"
)

// Country returns the ISO 3166-1 alpha-2 country code for the given IP address string.
// Returns an empty string if the IP is invalid, private, or the country is unknown.
func Country(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ""
	}

	country := iploc.IPCountry(addr)
	if country == "" || country == "ZZ" {
		return ""
	}

	return country
}
