package main

import (
	"testing"

	"github.com/sipeed/picoclaw/web/backend/launcherconfig"
)

func TestShouldEnableLauncherFileLogging(t *testing.T) {
	tests := []struct {
		name          string
		enableConsole bool
		debug         bool
		want          bool
	}{
		{name: "gui mode", enableConsole: false, debug: false, want: true},
		{name: "console mode", enableConsole: true, debug: false, want: false},
		{name: "debug gui mode", enableConsole: false, debug: true, want: true},
		{name: "debug console mode", enableConsole: true, debug: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldEnableLauncherFileLogging(tt.enableConsole, tt.debug); got != tt.want {
				t.Fatalf(
					"shouldEnableLauncherFileLogging(%t, %t) = %t, want %t",
					tt.enableConsole,
					tt.debug,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestDashboardTokenConfigHelpPath(t *testing.T) {
	const launcherPath = "/tmp/launcher-config.json"

	tests := []struct {
		name   string
		source launcherconfig.DashboardTokenSource
		want   string
	}{
		{
			name:   "env token does not expose config path",
			source: launcherconfig.DashboardTokenSourceEnv,
			want:   "",
		},
		{
			name:   "config token exposes config path",
			source: launcherconfig.DashboardTokenSourceConfig,
			want:   launcherPath,
		},
		{
			name:   "random token does not expose config path",
			source: launcherconfig.DashboardTokenSourceRandom,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dashboardTokenConfigHelpPath(tt.source, launcherPath); got != tt.want {
				t.Fatalf("dashboardTokenConfigHelpPath(%q, %q) = %q, want %q", tt.source, launcherPath, got, tt.want)
			}
		})
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Long token (>=12 chars): first 3 + 10 stars + last 4
		{"sdhjflsjdflksdf", "sdh**********ksdf"},
		{"abcdefghijklmnopqrstuvwxyz", "abc**********wxyz"},
		// Exactly 12 chars (3+4+5 hidden): suffix shown
		{"abcdefghijkl", "abc**********ijkl"},
		// 8 chars (minimum password length): suffix NOT shown — only prefix+stars
		{"abcdefgh", "abc**********"},
		// 11 chars (one below threshold): suffix NOT shown
		{"abcdefghijk", "abc**********"},
		// 4..3 chars: prefix shown, no suffix
		{"abcdefg", "abc**********"},
		{"abcd", "abc**********"},
		// <=3 chars: fully masked
		{"abc", "**********"},
		{"", "**********"},
	}
	for _, tt := range tests {
		if got := maskSecret(tt.input); got != tt.want {
			t.Errorf("maskSecret(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseLauncherHostList(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{name: "single host", raw: "127.0.0.1", want: []string{"127.0.0.1"}},
		{name: "multiple hosts", raw: "127.0.0.1, 192.168.2.5", want: []string{"127.0.0.1", "192.168.2.5"}},
		{name: "dedupe hosts", raw: "127.0.0.1,127.0.0.1", want: []string{"127.0.0.1"}},
		{name: "reject empty entry", raw: "127.0.0.1,  ", wantErr: true},
		{name: "reject empty input", raw: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLauncherHostList(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseLauncherHostList() err = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len(got) = %d, want %d (%#v)", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestResolveLauncherBindHost(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		envHost      string
		explicitHost bool
		effectivePub bool
		wantHost     string
		wantPublic   bool
		wantExplicit bool
		wantErr      bool
	}{
		{
			name:         "explicit host overrides public",
			host:         "0.0.0.0",
			explicitHost: true,
			effectivePub: true,
			wantHost:     "0.0.0.0",
			wantPublic:   false,
			wantExplicit: true,
		},
		{
			name:         "explicit host overrides env host",
			host:         "127.0.0.1",
			envHost:      "0.0.0.0",
			explicitHost: true,
			effectivePub: true,
			wantHost:     "127.0.0.1",
			wantPublic:   false,
			wantExplicit: true,
		},
		{
			name:         "explicit host cannot be empty",
			host:         "   ",
			explicitHost: true,
			effectivePub: false,
			wantErr:      true,
		},
		{
			name:         "env host overrides public",
			envHost:      "0.0.0.0",
			explicitHost: false,
			effectivePub: true,
			wantHost:     "0.0.0.0",
			wantPublic:   false,
			wantExplicit: true,
		},
		{
			name:         "explicit localhost uses adaptive private host",
			host:         "localhost",
			explicitHost: true,
			effectivePub: false,
			wantHost:     resolveDefaultLauncherPrivateHost(),
			wantPublic:   false,
			wantExplicit: true,
		},
		{
			name:         "explicit star uses adaptive any host",
			host:         "*",
			explicitHost: true,
			effectivePub: false,
			wantHost:     resolveDefaultLauncherAnyHost(),
			wantPublic:   false,
			wantExplicit: true,
		},
		{
			name:         "public mode without explicit host",
			host:         "",
			explicitHost: false,
			effectivePub: true,
			wantHost:     resolveDefaultLauncherAnyHost(),
			wantPublic:   true,
			wantExplicit: false,
		},
		{
			name:         "private mode without explicit host",
			host:         "",
			explicitHost: false,
			effectivePub: false,
			wantHost:     resolveDefaultLauncherPrivateHost(),
			wantPublic:   false,
			wantExplicit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPublic, gotExplicit, err := resolveLauncherBindHost(
				tt.host,
				tt.explicitHost,
				tt.envHost,
				tt.effectivePub,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveLauncherBindHost() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotHost != tt.wantHost {
				t.Fatalf("resolveLauncherBindHost() host = %q, want %q", gotHost, tt.wantHost)
			}
			if gotPublic != tt.wantPublic {
				t.Fatalf("resolveLauncherBindHost() public = %t, want %t", gotPublic, tt.wantPublic)
			}
			if gotExplicit != tt.wantExplicit {
				t.Fatalf("resolveLauncherBindHost() explicit = %t, want %t", gotExplicit, tt.wantExplicit)
			}
		})
	}
}

func TestResolveLauncherBindMode(t *testing.T) {
	tests := []struct {
		name         string
		rawHost      string
		hostExplicit bool
		effectivePub bool
		wantMode     launcherBindMode
	}{
		{name: "auto private", rawHost: "", hostExplicit: false, effectivePub: false, wantMode: launcherBindModeAutoPrivate},
		{name: "auto public", rawHost: "", hostExplicit: false, effectivePub: true, wantMode: launcherBindModeAutoPublic},
		{name: "explicit localhost", rawHost: "localhost", hostExplicit: true, effectivePub: false, wantMode: launcherBindModeExplicitAdaptiveLocal},
		{name: "explicit star", rawHost: "*", hostExplicit: true, effectivePub: false, wantMode: launcherBindModeExplicitAdaptiveAny},
		{name: "explicit literal", rawHost: "0.0.0.0", hostExplicit: true, effectivePub: false, wantMode: launcherBindModeExplicitLiteral},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveLauncherBindMode(tt.rawHost, tt.hostExplicit, tt.effectivePub); got != tt.wantMode {
				t.Fatalf("resolveLauncherBindMode() = %q, want %q", got, tt.wantMode)
			}
		})
	}
}

func TestLauncherConsoleHosts(t *testing.T) {
	t.Run("auto private includes dual loopback hints", func(t *testing.T) {
		hosts := launcherConsoleHosts(launcherBindModeAutoPrivate, "localhost", false)
		seen := make(map[string]bool, len(hosts))
		for _, host := range hosts {
			if seen[host] {
				t.Fatalf("duplicate host %q in %#v", host, hosts)
			}
			seen[host] = true
		}
		if !seen["localhost"] {
			t.Fatalf("expected localhost in %#v", hosts)
		}
		if !seen["::1"] {
			t.Fatalf("expected ::1 in %#v", hosts)
		}
		if !seen["127.0.0.1"] {
			t.Fatalf("expected 127.0.0.1 in %#v", hosts)
		}
	})

	t.Run("explicit ipv4 wildcard excludes ipv6 loopback", func(t *testing.T) {
		hosts := launcherConsoleHosts(launcherBindModeExplicitLiteral, "0.0.0.0", false)
		seen := make(map[string]bool, len(hosts))
		for _, host := range hosts {
			seen[host] = true
		}
		if seen["::1"] {
			t.Fatalf("did not expect ::1 in %#v", hosts)
		}
		if !seen["127.0.0.1"] {
			t.Fatalf("expected 127.0.0.1 in %#v", hosts)
		}
	})

	t.Run("explicit ipv6 host remains visible", func(t *testing.T) {
		hosts := launcherConsoleHosts(launcherBindModeExplicitLiteral, "::1", false)
		if len(hosts) != 2 {
			t.Fatalf("len(hosts) = %d, want 2 (%#v)", len(hosts), hosts)
		}
		if hosts[0] != "localhost" || hosts[1] != "::1" {
			t.Fatalf("hosts = %#v, want [localhost ::1]", hosts)
		}
	})
}

func TestBrowserHostForLauncher(t *testing.T) {
	if got := browserHostForLauncher("0.0.0.0"); got != "localhost" {
		t.Fatalf("browserHostForLauncher(0.0.0.0) = %q, want %q", got, "localhost")
	}
	if got := browserHostForLauncher("::"); got != "localhost" {
		t.Fatalf("browserHostForLauncher(::) = %q, want %q", got, "localhost")
	}
	if got := browserHostForLauncher("192.168.1.10"); got != "192.168.1.10" {
		t.Fatalf("browserHostForLauncher(192.168.1.10) = %q, want %q", got, "192.168.1.10")
	}
}

func TestWildcardAdvertiseIP(t *testing.T) {
	tests := []struct {
		name     string
		bindHost string
		ipv4     string
		ipv6     string
		want     string
	}{
		{name: "ipv4 wildcard prefers ipv6 when available", bindHost: "0.0.0.0", ipv4: "192.168.1.2", ipv6: "2001:db8::1", want: "2001:db8::1"},
		{name: "ipv6 wildcard uses ipv6", bindHost: "::", ipv4: "192.168.1.2", ipv6: "2001:db8::1", want: "2001:db8::1"},
		{name: "ipv6 wildcard falls back to ipv4", bindHost: "::", ipv4: "192.168.1.2", ipv6: "", want: "192.168.1.2"},
		{name: "ipv4 wildcard uses ipv6-only network", bindHost: "0.0.0.0", ipv4: "", ipv6: "2001:db8::1", want: "2001:db8::1"},
		{name: "non wildcard does not advertise", bindHost: "127.0.0.1", ipv4: "192.168.1.2", ipv6: "2001:db8::1", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wildcardAdvertiseIP(tt.bindHost, tt.ipv4, tt.ipv6); got != tt.want {
				t.Fatalf("wildcardAdvertiseIP(%q, %q, %q) = %q, want %q", tt.bindHost, tt.ipv4, tt.ipv6, got, tt.want)
			}
		})
	}
}
