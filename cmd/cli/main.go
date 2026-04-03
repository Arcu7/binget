package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Arcu7/binget/assets"
	"github.com/Arcu7/binget/internal/binary"
	"github.com/Arcu7/binget/internal/util/logger"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

func toStdErr(msg string, err ...error) {
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	fmt.Fprintf(os.Stderr, "%s: %s\n", red("ERROR"), msg)
	fmt.Fprintf(os.Stderr, "Details: %v\n", err)
	os.Exit(1)
}

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
		toStdErr("", err)
	} else {
		fmt.Println("Installation completed successfully!")
	}
}
