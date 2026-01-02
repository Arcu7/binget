package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/Arcu7/binget/assets"
	"github.com/Arcu7/binget/internal/binary"
	"github.com/Arcu7/binget/internal/util/logger"
	"github.com/Arcu7/binget/internal/util/syslist"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

func main() {
	args := setupArgs()
	flags := setupFlags()

	var log *slog.Logger

	cmd := &cli.Command{
		Name:                   "binget",
		Aliases:                []string{"bg", "bing", "bget"},
		Usage:                  "A simple CLI tool to install binaries from GitHub releases",
		UsageText:              "binget [global options] <repo>",
		HideVersion:            true,
		HideHelp:               true,
		UseShortOptionHandling: true,
		Arguments:              args,
		Flags:                  flags,
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			log = logger.New(cmd.Bool("verbose"))
			return nil, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println(assets.BingetASCII)
			fmt.Println("A simple CLI tool to install binaries from GitHub releases")
			fmt.Println("")

			repo := cmd.StringArg("repo")
			if repo == "" {
				cli.ShowRootCommandHelpAndExit(cmd, -1)
			}

			installMode := cmd.String("install-mode")
			authToken := cmd.String("token")

			goos, goarch, path, err := getAndCheckSystemRequirements()
			if err != nil {
				return err
			}

			path, err = getAndCheckCorrectPath(installMode, goos, path)
			if err != nil {
				return err
			}

			log.Info("System details", slog.String("os", goos), slog.String("arch", goarch), slog.String("path", path))
			finder := binary.NewFinder(log)

			err = finder.DownloadRelease(binary.Request{
				RepositoryURL: repo,
				AuthToken:     authToken,
				OS:            goos,
				Arch:          goarch,
				PathEnv:       path,
			})
			if err != nil {
				return err
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print binget's version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("binget version 1.0.0")
					return nil
				},
			},
			{
				Name:    "help",
				Aliases: []string{"h"},
				Usage:   "Print binget's help information",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cli.ShowRootCommandHelpAndExit(cmd, -1)
					return nil
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		red := color.New(color.FgRed, color.Bold).SprintFunc()
		fmt.Fprintf(os.Stderr, "%s: %v\n", red("ERROR"), err)
	} else {
		fmt.Println("Installation completed successfully!")
	}
}

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
		}
	case "system":
		if goos == syslist.OSLinux || goos == syslist.OSDarwin {
			for _, p := range paths {
				if p == syslist.PathSystemBin {
					return p, nil
				}
			}
			return "", fmt.Errorf("for 'system' install mode, ensure that %s is in your PATH environment variable", syslist.PathSystemBin)
		}
	default:
		return "", fmt.Errorf("unsupported install mode: %s", installMode)
	}

	return "", fmt.Errorf("unable to determine correct installation path for OS: %s", goos)
}
