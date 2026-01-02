// Package syslist lists out supported operating systems, architecture (and their aliases), and path variables.
// The supported OS and architecture is taken from Go internal/syslist package,
// Refer to https://go.dev/src/internal/syslist/syslist.go
package syslist

const (
	OSLinux  = "linux"
	OSDarwin = "darwin"

	ArchAMD64 = "amd64"
	Arch386   = "386"
	ArchARM64 = "arm64"

	PathUserBin   = ".local/bin"
	PathSystemBin = "/usr/local/bin"
)

var SupportedOS = map[string]struct{}{
	OSLinux:  {},
	OSDarwin: {},
	// TODO: Enable when Windows is supported
	// "windows":   struct{}{},
}

var SupportedArch = map[string]struct{}{
	ArchAMD64: {},
	Arch386:   {},
	ArchARM64: {},
}

var KnownArchAliases = map[string][]string{
	ArchAMD64: {"x86_64", "x86-64", "x64", "amd64"},
	Arch386:   {"i386", "i686", "32-bit", "386"},
	ArchARM64: {"aarch64", "arm64"},
}
