package utils

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

var (
	ipFamiliesOnce sync.Once
	hasIPv4        bool
	hasIPv6        bool
)

func DetectIPFamilies() (bool, bool) {
	ipFamiliesOnce.Do(func() {
		if ips, err := net.LookupIP("localhost"); err == nil {
			for _, ip := range ips {
				if ip == nil {
					continue
				}
				if ip.To4() != nil {
					hasIPv4 = true
					continue
				}
				hasIPv6 = true
			}
		}

		if hasIPv4 && hasIPv6 {
			return
		}

		if addrs, err := net.InterfaceAddrs(); err == nil {
			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if !ok || ipnet.IP == nil {
					continue
				}
				if ipnet.IP.To4() != nil {
					hasIPv4 = true
					continue
				}
				hasIPv6 = true
			}
		}
	})

	return hasIPv4, hasIPv6
}

func SelectAdaptiveLoopbackHost(hasIPv4, hasIPv6 bool) string {
	switch {
	case hasIPv4 && hasIPv6:
		return "localhost"
	case hasIPv6:
		return "::1"
	case hasIPv4:
		return "127.0.0.1"
	default:
		return "localhost"
	}
}

func SelectAdaptiveAnyHost(hasIPv4, hasIPv6 bool) string {
	switch {
	case hasIPv4 && hasIPv6:
		return "::"
	case hasIPv6:
		return "::"
	case hasIPv4:
		return "0.0.0.0"
	default:
		return "::"
	}
}

func ResolveAdaptiveLoopbackHost() string {
	hasIPv4, hasIPv6 := DetectIPFamilies()
	return SelectAdaptiveLoopbackHost(hasIPv4, hasIPv6)
}

func ResolveAdaptiveAnyHost() string {
	hasIPv4, hasIPv6 := DetectIPFamilies()
	return SelectAdaptiveAnyHost(hasIPv4, hasIPv6)
}

// GetPicoclawHome returns the picoclaw home directory.
// Priority: $PICOCLAW_HOME > ~/.picoclaw
func GetPicoclawHome() string {
	return config.GetHome()
}

// GetDefaultConfigPath returns the default path to the picoclaw config file.
func GetDefaultConfigPath() string {
	if configPath := os.Getenv(config.EnvConfig); configPath != "" {
		return configPath
	}
	return filepath.Join(GetPicoclawHome(), "config.json")
}

// FindPicoclawBinary locates the picoclaw executable.
// Search order:
//  1. PICOCLAW_BINARY environment variable (explicit override)
//  2. Same directory as the current executable
//  3. Falls back to "picoclaw" and relies on $PATH
func FindPicoclawBinary() string {
	binaryName := "picoclaw"
	if runtime.GOOS == "windows" {
		binaryName = "picoclaw.exe"
	}

	if p := os.Getenv(config.EnvBinary); p != "" {
		if info, _ := os.Stat(p); info != nil && !info.IsDir() {
			return p
		}
	}

	if exe, err := os.Executable(); err == nil {
		logger.Debugf("Trying to find picoclaw binary in %s", exe)
		candidate := filepath.Join(filepath.Dir(exe), binaryName)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}

	return "picoclaw"
}

// GetLocalIPv4 returns a non-loopback local IPv4 address.
func GetLocalIPv4() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

// GetLocalIPv6 returns a non-loopback local IPv6 address.
func GetLocalIPv6() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP == nil {
			continue
		}
		ip := ipnet.IP
		if ip.IsLoopback() || ip.To4() != nil {
			continue
		}
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			continue
		}
		return ip.String()
	}
	return ""
}

// GetLocalIP returns a non-loopback local IPv4 address for backward compatibility.
func GetLocalIP() string {
	return GetLocalIPv4()
}

// OpenBrowser automatically opens the given URL in the default browser.
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
