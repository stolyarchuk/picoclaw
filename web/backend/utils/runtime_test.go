package utils

import "testing"

func TestSelectAdaptiveLoopbackHost(t *testing.T) {
	tests := []struct {
		name    string
		hasIPv4 bool
		hasIPv6 bool
		want    string
	}{
		{name: "dual stack", hasIPv4: true, hasIPv6: true, want: "localhost"},
		{name: "ipv6 only", hasIPv4: false, hasIPv6: true, want: "::1"},
		{name: "ipv4 only", hasIPv4: true, hasIPv6: false, want: "127.0.0.1"},
		{name: "fallback", hasIPv4: false, hasIPv6: false, want: "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectAdaptiveLoopbackHost(tt.hasIPv4, tt.hasIPv6); got != tt.want {
				t.Fatalf("SelectAdaptiveLoopbackHost(%t, %t) = %q, want %q", tt.hasIPv4, tt.hasIPv6, got, tt.want)
			}
		})
	}
}

func TestSelectAdaptiveAnyHost(t *testing.T) {
	tests := []struct {
		name    string
		hasIPv4 bool
		hasIPv6 bool
		want    string
	}{
		{name: "dual stack", hasIPv4: true, hasIPv6: true, want: "::"},
		{name: "ipv6 only", hasIPv4: false, hasIPv6: true, want: "::"},
		{name: "ipv4 only", hasIPv4: true, hasIPv6: false, want: "0.0.0.0"},
		{name: "fallback", hasIPv4: false, hasIPv6: false, want: "::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectAdaptiveAnyHost(tt.hasIPv4, tt.hasIPv6); got != tt.want {
				t.Fatalf("SelectAdaptiveAnyHost(%t, %t) = %q, want %q", tt.hasIPv4, tt.hasIPv6, got, tt.want)
			}
		})
	}
}

func TestResolveAdaptiveHosts(t *testing.T) {
	loopback := ResolveAdaptiveLoopbackHost()
	if loopback == "" {
		t.Fatal("ResolveAdaptiveLoopbackHost() returned empty host")
	}

	anyHost := ResolveAdaptiveAnyHost()
	if anyHost == "" {
		t.Fatal("ResolveAdaptiveAnyHost() returned empty host")
	}
}
