package auth

import (
	"fmt"
	"net"
)

// CanonicalizeIP converts an IP address to its canonical 16-byte string representation.
// This ensures consistent storage and comparison regardless of input format.
// For example, "2001:db8::1" and "2001:db8:0:0:0:0:0:1" will both produce the same output.
func CanonicalizeIP(ip string) (string, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}

	// Convert to 16-byte representation for consistency
	// IPv4 addresses will be represented as IPv4-mapped IPv6 addresses
	canonical := parsed.To16()
	if canonical == nil {
		return "", fmt.Errorf("failed to canonicalize IP address: %s", ip)
	}

	return canonical.String(), nil
}

// CanonicalizeIPs converts a slice of IP addresses to their canonical forms.
// Returns an error if any IP is invalid.
func CanonicalizeIPs(ips []string) ([]string, error) {
	result := make([]string, len(ips))
	for i, ip := range ips {
		canonical, err := CanonicalizeIP(ip)
		if err != nil {
			return nil, err
		}
		result[i] = canonical
	}
	return result, nil
}

// IsIPAllowed checks if the given IP is in the allowed list.
// If the allowed list is empty, all IPs are allowed.
// Both the input IP and the allowed list should already be canonicalized.
func IsIPAllowed(ip string, allowedIPs []string) bool {
	if len(allowedIPs) == 0 {
		return true
	}

	for _, allowed := range allowedIPs {
		if ip == allowed {
			return true
		}
	}
	return false
}

// ValidateAndCanonicalizeIP validates an IP address and returns its canonical form.
// This is a convenience function combining validation and canonicalization.
func ValidateAndCanonicalizeIP(ip string) (string, bool) {
	canonical, err := CanonicalizeIP(ip)
	if err != nil {
		return "", false
	}
	return canonical, true
}
