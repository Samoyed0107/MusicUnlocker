package outputplan

import (
	"errors"
	"path/filepath"
	"testing"

	"musicunlocker/internal/probe"
)

func TestPlanCleanNames(t *testing.T) {
	tests := []struct {
		name, input, format string
		family              probe.Family
		want                string
	}{
		{"kgm", "foo.kgm", "mp3", probe.FamilyKugou, "foo.mp3"},
		{"kgma", "foo.kgma", "flac", probe.FamilyKugou, "foo.flac"},
		{"vpr", "foo.vpr", "mp3", probe.FamilyKugou, "foo.mp3"},
		{"compound", "foo.kgm.flac", "mp3", probe.FamilyKugou, "foo.mp3"},
		{"counter", "foo.kgm(1).flac", "mp3", probe.FamilyKugou, "foo.mp3"},
		{"counter spaced", "foo.kgm (1).flac", "mp3", probe.FamilyKugou, "foo.mp3"},
		{"ncm mp3", "foo.ncm", "mp3", probe.FamilyNetease, "foo.mp3"},
		{"ncm flac", "foo.ncm", "flac", probe.FamilyNetease, "foo.flac"},
		{"kwm mp3", "foo.kwm", "mp3", probe.FamilyKuwo, "foo.mp3"},
		{"kwm flac", "foo.kwm", "flac", probe.FamilyKuwo, "foo.flac"},
		{"mgg counter", "foo.mgg0(1).flac", "ogg", probe.FamilyQMC, "foo.ogg"},
		{"mflac", "foo.mflac0.flac", "flac", probe.FamilyQMC, "foo.flac"},
		{"title parentheses", "Artist - Song (Live).ncm", "mp3", probe.FamilyNetease, "Artist - Song (Live).mp3"},
		{"remaster parentheses", "Artist - Song (2024 Remaster).kwm", "flac", probe.FamilyKuwo, "Artist - Song (2024 Remaster).flac"},
		{"misleading jpg", "foo.jpg", "mp3", probe.FamilyKugou, "foo.mp3"},
		{"misleading flac", "foo.flac", "mp3", probe.FamilyKugou, "foo.mp3"},
		{"extensionless", "foo", "mp3", probe.FamilyKugou, "foo.mp3"},
		{"tm", "foo.tm2", "m4a", probe.FamilyTM, "foo.m4a"},
		{"ximalaya", "foo.x2m", "mp3", probe.FamilyXimalaya, "foo.mp3"},
		{"ximalaya m4a", "foo.x3m", "m4a", probe.FamilyXimalaya, "foo.m4a"},
		{"xiami", "foo.xm", "flac", probe.FamilyXiami, "foo.flac"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Plan(Request{InputPath: tc.input, OutputDir: `D:\out`, Probe: probe.Result{Family: tc.family}, DecodedFormat: tc.format})
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(got) != tc.want {
				t.Fatalf("got %q, want %q", filepath.Base(got), tc.want)
			}
		})
	}
}

func TestPlanSourceConflict(t *testing.T) {
	_, err := Plan(Request{InputPath: `D:\out\foo.mp3`, OutputDir: `D:\out`, Probe: probe.Result{Family: probe.FamilyKugou}, DecodedFormat: "mp3"})
	if !errors.Is(err, ErrSourceConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestTargetKeyOmitsDecodedExtension(t *testing.T) {
	kgm, err := TargetKey(Request{InputPath: "foo.kgm", OutputDir: `D:\out`, Probe: probe.Result{Family: probe.FamilyKugou}})
	if err != nil {
		t.Fatal(err)
	}
	kgma, err := TargetKey(Request{InputPath: "foo.kgma", OutputDir: `D:\out`, Probe: probe.Result{Family: probe.FamilyKugou}})
	if err != nil {
		t.Fatal(err)
	}
	if kgm != kgma {
		t.Fatalf("keys differ: %q vs %q", kgm, kgma)
	}
}
