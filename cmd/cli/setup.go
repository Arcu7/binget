package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Arcu7/binget/internal/util/syslist"
	"github.com/urfave/cli/v3"
)

func setupArgs() []cli.Argument {
	var args []cli.Argument

	repoURLArg := &cli.StringArg{
		Name:      "repo",
		UsageText: "GitHub repository URL (e.g.,)",
	}

	args = append(args, repoURLArg)

	return args
}

func setupFlags() []cli.Flag {
	var flags []cli.Flag

	installModeFlag := &cli.StringFlag{
		Name:        "install-mode",
		Aliases:     []string{"m"},
		Value:       "user",
		Usage:       "Installation mode: 'user' or 'system'",
		DefaultText: "user",
		Action: func(ctx context.Context, cmd *cli.Command, mode string) error {
			if mode != "user" && mode != "system" {
				return fmt.Errorf("invalid install mode: %s. Valid options are 'user' or 'system'", mode)
			}
			return nil
		},
	}

	// Token flag for GitHub authentication (optional, but needed for increased rate limits)
	tokenFlag := &cli.StringFlag{
		Name:        "token",
		Aliases:     []string{"t"},
		Value:       "",
		Usage:       "Authentication token for GitHub API requests (optional)",
		Sources:     cli.EnvVars("BINGET_GITHUB_TOKEN"),
		DefaultText: "\"\" or env:BINGET_GITHUB_TOKEN",
	}

	verboseFlag := &cli.BoolFlag{
		Name:        "verbose",
		Aliases:     []string{"v"},
		Value:       false,
		Usage:       "Enable verbose output",
		DefaultText: "false",
	}

	flags = append(flags, installModeFlag, tokenFlag, verboseFlag)

	return flags
}

func getAndCheckSystemRequirements() (goos, goarch, path string, err error) {
	_, exist := syslist.SupportedOS[runtime.GOOS]
	if !exist {
		if runtime.GOOS == "windows" {
			return goos, goarch, path, fmt.Errorf("binget is currently not supported on Windows")
		}
		return goos, goarch, path, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	goos = runtime.GOOS

	_, exist = syslist.SupportedArch[runtime.GOARCH]
	if !exist {
		return goos, goarch, path, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	goarch = runtime.GOARCH

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return goos, goarch, pathEnv, fmt.Errorf("PATH environment variable is not set")
	}

	return goos, goarch, pathEnv, nil
}

func getAndCheckCorrectPath(installMode, goos, pathEnv string) (string, error) {
	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	switch installMode {
	case "user":
		homeDir, err := os.UserHomeDir()
		if goos == syslist.OSLinux || goos == syslist.OSDarwin {
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			userBinPath := fmt.Sprintf("%s/%s", homeDir, syslist.PathUserBin)

			for _, p := range paths {
				if p == userBinPath {
					return p, nil
				}
			}

			return "", fmt.Errorf("for 'user' install mode, ensure that %s is in your PATH environment variable", syslist.PathUserBin)
		} else {
			return "", fmt.Errorf("unsupported OS for 'user' install mode: %s", goos)
		}
	case "system":
		if goos == syslist.OSLinux || goos == syslist.OSDarwin {
			for _, p := range paths {
				if p == syslist.PathSystemBin {
					return p, nil
				}
			}
			return "", fmt.Errorf("for 'system' install mode, ensure that %s is in your PATH environment variable", syslist.PathSystemBin)
		} else {
			return "", fmt.Errorf("unsupported OS for 'system' install mode: %s", goos)
		}
	default:
		return "", fmt.Errorf("unsupported install mode: %s", installMode)
	}
}
