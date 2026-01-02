// Package assets contains embedded asset files.
// The embedding is done at the root level since go embed doesnt' support
// embedding from subdirectories directly.
package assets

import _ "embed"

// BingetASCII is to store the ASCII art displayed when running the CLI tool.
// TODO: Maybe explore not using embed for this and just have it as a constant string
// Problem is:
// 1. The ASCII art contains backticks (should be escapable but need more research)
//
//go:embed asciiArt.txt
var BingetASCII string
