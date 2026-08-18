package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"musicunlocker/internal/probe"
)

type fileReport struct {
	Path   string      `json:"path"`
	Size   int64       `json:"size"`
	Result *jsonResult `json:"result"`
	Error  string      `json:"error,omitempty"`
}

type report struct {
	Files []fileReport `json:"files"`
}

type jsonResult struct {
	Family           probe.Family     `json:"family"`
	Variant          string           `json:"variant"`
	Confidence       probe.Confidence `json:"confidence"`
	RawAudioFormat   string           `json:"rawAudioFormat"`
	FooterType       string           `json:"footerType"`
	NeedsExternalKey bool             `json:"needsExternalKey"`
	AudioOffset      uint32           `json:"audioOffset"`
	CryptoVersion    uint32           `json:"cryptoVersion"`
	CryptoSlot       uint32           `json:"cryptoSlot"`
	TagVersion       uint32           `json:"tagVersion"`
	TagSize          uint32           `json:"tagSize"`
	ExtensionHints   []string         `json:"extensionHints"`
	Evidence         []probe.Evidence `json:"evidence"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("musicprobe", flag.ContinueOnError)
	flags.SetOutput(errOut)
	jsonMode := flags.Bool("json", false, "emit one JSON report containing all files")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	paths := flags.Args()
	if len(paths) == 0 {
		fmt.Fprintln(errOut, "usage: musicprobe [--json] <file> [file ...]")
		return 2
	}

	report := report{Files: make([]fileReport, 0, len(paths))}
	exitCode := 0
	for _, path := range paths {
		entry := fileReport{Path: path, Size: -1}
		info, statErr := os.Stat(path)
		if statErr != nil {
			entry.Error = statErr.Error()
			exitCode = 1
		} else {
			entry.Size = info.Size()
			result, probeErr := probe.ProbeFile(path)
			if probeErr != nil {
				entry.Error = probeErr.Error()
				exitCode = 1
			} else {
				entry.Result = toJSONResult(result)
			}
		}
		report.Files = append(report.Files, entry)
	}

	if *jsonMode {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
		return exitCode
	}
	for i, entry := range report.Files {
		if i > 0 {
			fmt.Fprintln(out)
		}
		writeHuman(out, entry)
	}
	return exitCode
}

func toJSONResult(result probe.Result) *jsonResult {
	return &jsonResult{
		Family: result.Family, Variant: result.Variant, Confidence: result.Confidence,
		RawAudioFormat: result.RawAudioFormat, FooterType: result.FooterType,
		NeedsExternalKey: result.NeedsExternalKey, AudioOffset: result.AudioOffset,
		CryptoVersion: result.CryptoVersion, CryptoSlot: result.CryptoSlot,
		TagVersion: result.TagVersion, TagSize: result.TagSize,
		ExtensionHints: result.ExtensionHints, Evidence: result.Evidence,
	}
}

func writeHuman(out io.Writer, entry fileReport) {
	fmt.Fprintf(out, "Path: %s\nSize: %d\n", entry.Path, entry.Size)
	if entry.Error != "" {
		fmt.Fprintf(out, "Error: %s\n", entry.Error)
		return
	}
	result := entry.Result
	fmt.Fprintf(out, "Family: %s\nVariant: %s\nConfidence: %s\n", result.Family, result.Variant, result.Confidence)
	fmt.Fprintf(out, "RawAudioFormat: %s\nFooterType: %s\nNeedsExternalKey: %t\n", unavailable(result.RawAudioFormat), unavailable(result.FooterType), result.NeedsExternalKey)
	fmt.Fprintf(out, "AudioOffset: %s\nCryptoVersion: %s\nCryptoSlot: %s\n", number(result.AudioOffset, result.Family == probe.FamilyKugou), number(result.CryptoVersion, result.Family == probe.FamilyKugou), number(result.CryptoSlot, result.Family == probe.FamilyKugou))
	fmt.Fprintf(out, "ExtensionHints:\n%sEvidence:\n%s", list(result.ExtensionHints), evidence(result.Evidence))
}

func unavailable(value string) string {
	if value == "" {
		return "unavailable"
	}
	return value
}

func number(value uint32, meaningful bool) string {
	if !meaningful {
		return "unavailable"
	}
	return fmt.Sprint(value)
}

func list(values []string) string {
	if len(values) == 0 {
		return "  - unavailable\n"
	}
	var b strings.Builder
	for _, value := range values {
		fmt.Fprintf(&b, "  - %s\n", value)
	}
	return b.String()
}

func evidence(values []probe.Evidence) string {
	if len(values) == 0 {
		return "  - unavailable\n"
	}
	var b strings.Builder
	for _, value := range values {
		fmt.Fprintf(&b, "  - %s: %s\n", value.Kind, value.Detail)
	}
	return b.String()
}
