package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"musicunlocker/internal/decode"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("musicdecode", flag.ContinueOnError)
	flags.SetOutput(errOut)
	outDir := flags.String("out-dir", "", "explicit output directory")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *outDir == "" || flags.NArg() != 1 {
		fmt.Fprintln(errOut, "usage: musicdecode --out-dir <dir> <file>")
		return 2
	}
	result, err := decode.DecodeFile(context.Background(), flags.Arg(0), decode.Options{OutDir: *outDir})
	if err != nil {
		if code := decode.CodeOf(err); code != "" {
			fmt.Fprintf(errOut, "%s: %v\n", code, err)
		} else {
			fmt.Fprintln(errOut, err)
		}
		return 1
	}
	fmt.Fprintf(out, "Input: %s\nEncryptedFamily: %s\nEncryptedVariant: %s\nCryptoVersion: %d\nDecodedFormat: %s\nOutput: %s\nBytesWritten: %d\nStatus: OK\n", result.InputPath, result.EncryptedFamily, result.EncryptedVariant, result.CryptoVersion, result.DecodedFormat, result.OutputPath, result.BytesWritten)
	return 0
}
