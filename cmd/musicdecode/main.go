package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"musicunlocker/internal/decode"
	"musicunlocker/internal/naming"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("musicdecode", flag.ContinueOnError)
	flags.SetOutput(errOut)
	outDir := flags.String("out-dir", "", "explicit output directory")
	nameMode := flags.String("name-mode", "source", "source, title, artist-title, title-artist, track-title, artist-album-track-title")
	nameTemplate := flags.String("name-template", "", "custom output-name template")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *outDir == "" || flags.NArg() != 1 {
		fmt.Fprintln(errOut, "usage: musicdecode --out-dir <dir> <file>")
		return 2
	}
	config, err := parseNaming(*nameMode, *nameTemplate)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 2
	}
	result, err := decode.DecodeFile(context.Background(), flags.Arg(0), decode.Options{OutDir: *outDir, Naming: config})
	if err != nil {
		writeError(errOut, err)
		return 1
	}
	fmt.Fprintf(out, "Input: %s\nEncryptedFamily: %s\nEncryptedVariant: %s\nCryptoVersion: %d\nDecodedFormat: %s\nOutput: %s\nBytesWritten: %d\nStatus: OK\n", result.InputPath, result.EncryptedFamily, result.EncryptedVariant, result.CryptoVersion, result.DecodedFormat, result.OutputPath, result.BytesWritten)
	return 0
}

func parseNaming(mode, template string) (naming.Config, error) {
	if template != "" {
		return naming.Config{Mode: naming.Custom, CustomTemplate: template}, nil
	}
	var m naming.Mode
	switch mode {
	case "", "source":
		m = naming.KeepSource
	case "title":
		m = naming.Title
	case "artist-title":
		m = naming.ArtistTitle
	case "title-artist":
		m = naming.TitleArtist
	case "track-title":
		m = naming.TrackTitle
	case "artist-album-track-title":
		m = naming.ArtistAlbumTrackTitle
	default:
		return naming.Config{}, fmt.Errorf("unknown --name-mode %q", mode)
	}
	return naming.Config{Mode: m}, naming.Validate(naming.Config{Mode: m})
}

func writeError(out io.Writer, err error) { fmt.Fprintln(out, err) }
