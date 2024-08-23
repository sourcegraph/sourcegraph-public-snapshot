// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package confignet // import "go.opentelemetry.io/collector/config/confignet"

import (
	"context"
	"fmt"
	"net"
	"time"
)

// TransportType represents a type of network transport protocol
type TransportType string

const (
	TransportTypeTCP        TransportType = "tcp"
	TransportTypeTCP4       TransportType = "tcp4"
	TransportTypeTCP6       TransportType = "tcp6"
	TransportTypeUDP        TransportType = "udp"
	TransportTypeUDP4       TransportType = "udp4"
	TransportTypeUDP6       TransportType = "udp6"
	TransportTypeIP         TransportType = "ip"
	TransportTypeIP4        TransportType = "ip4"
	TransportTypeIP6        TransportType = "ip6"
	TransportTypeUnix       TransportType = "unix"
	TransportTypeUnixgram   TransportType = "unixgram"
	TransportTypeUnixPacket TransportType = "unixpacket"
	transportTypeEmpty      TransportType = ""
)

// UnmarshalText unmarshalls text to a TransportType.
// Valid values are "tcp", "tcp4", "tcp6", "udp", "udp4",
// "udp6", "ip", "ip4", "ip6", "unix", "unixgram" and "unixpacket"
func (tt *TransportType) UnmarshalText(in []byte) error {
	typ := TransportType(in)
	switch typ {
	case TransportTypeTCP,
		TransportTypeTCP4,
		TransportTypeTCP6,
		TransportTypeUDP,
		TransportTypeUDP4,
		TransportTypeUDP6,
		TransportTypeIP,
		TransportTypeIP4,
		TransportTypeIP6,
		TransportTypeUnix,
		TransportTypeUnixgram,
		TransportTypeUnixPacket,
		transportTypeEmpty:
		*tt = typ
		return nil
	default:
		return fmt.Errorf("unsupported transport type %q", typ)
	}
}

// DialerConfig contains options for connecting to an address.
type DialerConfig struct {
	// Timeout is the maximum amount of time a dial will wait for
	// a connect to complete. The default is no timeout.
	Timeout time.Duration `mapstructure:"timeout"`
}

// NewDefaultDialerConfig creates a new DialerConfig with any default values set
func NewDefaultDialerConfig() DialerConfig {
	return DialerConfig{}
}

// AddrConfig represents a network endpoint address.
type AddrConfig struct {
	// Endpoint configures the address for this network connection.
	// For TCP and UDP networks, the address has the form "host:port". The host must be a literal IP address,
	// or a host name that can be resolved to IP addresses. The port must be a literal port number or a service name.
	// If the host is a literal IPv6 address it must be enclosed in square brackets, as in "[2001:db8::1]:80" or
	// "[fe80::1%zone]:80". The zone specifies the scope of the literal IPv6 address as defined in RFC 4007.
	Endpoint string `mapstructure:"endpoint"`

	// Transport to use. Allowed protocols are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only), "udp", "udp4" (IPv4-only),
	// "udp6" (IPv6-only), "ip", "ip4" (IPv4-only), "ip6" (IPv6-only), "unix", "unixgram" and "unixpacket".
	Transport TransportType `mapstructure:"transport"`

	// DialerConfig contains options for connecting to an address.
	DialerConfig DialerConfig `mapstructure:"dialer"`
}

// NewDefaultAddrConfig creates a new AddrConfig with any default values set
func NewDefaultAddrConfig() AddrConfig {
	return AddrConfig{
		DialerConfig: NewDefaultDialerConfig(),
	}
}

// Dial equivalent with net.Dialer's DialContext for this address.
func (na *AddrConfig) Dial(ctx context.Context) (net.Conn, error) {
	d := net.Dialer{Timeout: na.DialerConfig.Timeout}
	return d.DialContext(ctx, string(na.Transport), na.Endpoint)
}

// Listen equivalent with net.ListenConfig's Listen for this address.
func (na *AddrConfig) Listen(ctx context.Context) (net.Listener, error) {
	lc := net.ListenConfig{}
	return lc.Listen(ctx, string(na.Transport), na.Endpoint)
}

func (na *AddrConfig) Validate() error {
	switch na.Transport {
	case TransportTypeTCP,
		TransportTypeTCP4,
		TransportTypeTCP6,
		TransportTypeUDP,
		TransportTypeUDP4,
		TransportTypeUDP6,
		TransportTypeIP,
		TransportTypeIP4,
		TransportTypeIP6,
		TransportTypeUnix,
		TransportTypeUnixgram,
		TransportTypeUnixPacket:
		return nil
	default:
		return fmt.Errorf("invalid transport type %q", na.Transport)
	}
}

// TCPAddrConfig represents a TCP endpoint address.
type TCPAddrConfig struct {
	// Endpoint configures the address for this network connection.
	// The address has the form "host:port". The host must be a literal IP address, or a host name that can be
	// resolved to IP addresses. The port must be a literal port number or a service name.
	// If the host is a literal IPv6 address it must be enclosed in square brackets, as in "[2001:db8::1]:80" or
	// "[fe80::1%zone]:80". The zone specifies the scope of the literal IPv6 address as defined in RFC 4007.
	Endpoint string `mapstructure:"endpoint"`

	// DialerConfig contains options for connecting to an address.
	DialerConfig DialerConfig `mapstructure:"dialer"`
}

// NewDefaultTCPAddrConfig creates a new TCPAddrConfig with any default values set
func NewDefaultTCPAddrConfig() TCPAddrConfig {
	return TCPAddrConfig{
		DialerConfig: NewDefaultDialerConfig(),
	}
}

// Dial equivalent with net.Dialer's DialContext for this address.
func (na *TCPAddrConfig) Dial(ctx context.Context) (net.Conn, error) {
	d := net.Dialer{Timeout: na.DialerConfig.Timeout}
	return d.DialContext(ctx, string(TransportTypeTCP), na.Endpoint)
}

// Listen equivalent with net.ListenConfig's Listen for this address.
func (na *TCPAddrConfig) Listen(ctx context.Context) (net.Listener, error) {
	lc := net.ListenConfig{}
	return lc.Listen(ctx, string(TransportTypeTCP), na.Endpoint)
}
