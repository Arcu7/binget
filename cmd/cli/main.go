package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Arcu7/binget/internal/binary"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

func main() {
	args := setupArgs()
	flags := setupFlags()

	cmd := &cli.Command{
		Name:      "binget",
		Aliases:   []string{"bg", "bing", "bget"},
		Usage:     "A simple CLI tool to install binaries from GitHub releases",
		Version:   "v1.0.0",
		Arguments: args,
		Flags:     flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			repo := cmd.StringArg("repo")
			if repo == "" {
				cli.ShowRootCommandHelpAndExit(cmd, -1)
			}

			err := binary.DownloadRelease(binary.Request{
				RepositoryURL: cmd.StringArg("repo"),
				AuthToken:     cmd.String("token"),
			})
			if err != nil {
				return err
			}
			return nil
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		red := color.New(color.FgRed, color.Bold).SprintFunc()
		fmt.Fprintf(os.Stderr, "%s: %v\n", red("ERROR"), err)
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
	// Token flag for GitHub authentication (optional, but needed for increased rate limits)
	tokenFlag := &cli.StringFlag{
		Name:    "token",
		Aliases: []string{"t"},
		Value:   "",
		Usage:   "GitHub PAT token for authentication",
		Sources: cli.EnvVars("BINGET_GITHUB_TOKEN"),
	}

	flags = append(flags, tokenFlag)

	return flags
}
