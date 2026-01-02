// Package assets contains embedded asset files.
// The embedding is done at the root leve since go embed doesnt' support
// embedding from subdirectories directly.
package assets

import _ "embed"

//go:embed asciiArt.txt
var BingetASCII string
