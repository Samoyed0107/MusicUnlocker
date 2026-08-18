package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"musicunlocker/internal/batch"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("musicbatch", flag.ContinueOnError)
	flags.SetOutput(errOut)
	outDir := flags.String("out-dir", "", "explicit output directory")
	jsonOutput := flags.Bool("json", false, "write the structured summary as JSON")
	jobs := flags.Int("jobs", 1, "maximum concurrent decode workers")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *outDir == "" || flags.NArg() == 0 || *jobs <= 0 {
		fmt.Fprintln(errOut, "usage: musicbatch --out-dir <dir> [--jobs N] [--json] <file-or-directory> [more...]")
		return 2
	}
	summary, err := (batch.Engine{}).Run(context.Background(), batch.Request{Inputs: flags.Args(), OutputRoot: *outDir, MaxWorkers: *jobs})
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 2
	}
	if *jsonOutput {
		if err := json.NewEncoder(out).Encode(summary); err != nil {
			fmt.Fprintln(errOut, err)
			return 2
		}
	} else {
		for _, result := range summary.Results {
			printResult(out, result)
		}
		fmt.Fprintf(out, "\nSummary:\nSuccess: %d\nSkipped: %d\nBlocked: %d\nUnsupported: %d\nUnknown: %d\nFailed: %d\nCancelled: %d\nBytes written: %d\n", summary.Successes, summary.SkippedNotEncrypted+summary.SkippedExisting, summary.BlockedExternalKey+summary.BlockedExternalDatabase, summary.Unsupported, summary.Unknown, summary.Failed, summary.Cancelled, summary.BytesWritten)
	}
	if summary.Cancelled > 0 {
		return 3
	}
	if summary.Failed > 0 {
		return 1
	}
	return 0
}

func printResult(out io.Writer, r batch.Result) {
	name := filepath.Base(r.SourcePath)
	switch r.Status {
	case batch.StatusSuccess:
		fmt.Fprintf(out, "[OK] %s -> %s\n", name, filepath.Base(r.OutputPath))
	case batch.StatusSkippedNotEncrypted:
		fmt.Fprintf(out, "[SKIP] %s : not encrypted\n", name)
	case batch.StatusSkippedOutputExists:
		fmt.Fprintf(out, "[EXISTS] %s -> %s\n", name, filepath.Base(r.OutputPath))
	case batch.StatusBlockedExternalKey:
		fmt.Fprintf(out, "[BLOCKED_KEY] %s : %s\n", name, r.ErrorDetail)
	case batch.StatusBlockedExternalDatabase:
		fmt.Fprintf(out, "[BLOCKED_DATABASE] %s : %s\n", name, r.ErrorDetail)
	case batch.StatusUnsupported:
		fmt.Fprintf(out, "[UNSUPPORTED] %s : %s\n", name, r.ErrorDetail)
	case batch.StatusUnknown:
		fmt.Fprintf(out, "[UNKNOWN] %s : %s\n", name, r.ErrorDetail)
	case batch.StatusCancelled:
		fmt.Fprintf(out, "[CANCELLED] %s : %s\n", name, r.ErrorDetail)
	case batch.StatusFailed:
		if r.ErrorCode == batch.CodeDiscoveryError {
			fmt.Fprintf(out, "[DISCOVERY_ERROR] %s : %s\n", name, r.ErrorDetail)
			return
		}
		fmt.Fprintf(out, "[FAIL] %s : %s\n", name, r.ErrorDetail)
	default:
		fmt.Fprintf(out, "[FAIL] %s : %s\n", name, r.ErrorDetail)
	}
}
