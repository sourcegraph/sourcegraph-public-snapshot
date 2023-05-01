package requestclient

import (
	"net"
	"testing"
)

func TestBaseIP(t *testing.T) {
	tests := []struct {
		name string
		addr net.Addr
		want string
	}{
		{
			name: "TCP address",
			addr: &net.TCPAddr{
				IP:   net.ParseIP("127.0.127.2"),
				Port: 448,
			},
			want: "127.0.127.2",
		},
		{
			name: "UDP address",
			addr: &net.UDPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 448,
			},
			want: "127.0.0.1",
		},
		{
			name: "Other address",
			addr: &net.UnixAddr{
				Name: "foobar",
			},
			want: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := baseIP(tt.addr); got != tt.want {
				t.Errorf("baseIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
