package upstream

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	canonicalqmc "git.um-react.app/um/cli/algo/qmc"
	_ "modernc.org/sqlite"
)

func syntheticKGMV3() []byte {
	b := make([]byte, 60+4)
	copy(b, []byte{0x7c, 0xd5, 0x32, 0xeb, 0x86, 0x02, 0x7f, 0x4b, 0xa8, 0xaf, 0xa6, 0x8e, 0x0f, 0xff, 0x99, 0x14})
	binary.LittleEndian.PutUint32(b[0x10:], 60)
	binary.LittleEndian.PutUint32(b[0x14:], 3)
	binary.LittleEndian.PutUint32(b[0x18:], 1)
	return b
}

func syntheticKGMV5(audioHash string) []byte {
	b := make([]byte, 72+len(audioHash))
	copy(b, []byte{0x7c, 0xd5, 0x32, 0xeb, 0x86, 0x02, 0x7f, 0x4b, 0xa8, 0xaf, 0xa6, 0x8e, 0x0f, 0xff, 0x99, 0x14})
	binary.LittleEndian.PutUint32(b[0x10:], 60)
	binary.LittleEndian.PutUint32(b[0x14:], 5)
	binary.LittleEndian.PutUint32(b[0x18:], 1)
	binary.LittleEndian.PutUint32(b[68:], uint32(len(audioHash)))
	copy(b[72:], audioHash)
	return b
}

func makeSyntheticKGGDatabase(t *testing.T, keyID, key string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kgg.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`create table ShareFileItems (EncryptionKeyId text, EncryptionKey text)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`insert into ShareFileItems (EncryptionKeyId, EncryptionKey) values (?, ?)`, keyID, key); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDecodeKGMV5DatabaseLookupBoundary(t *testing.T) {
	dbPath := makeSyntheticKGGDatabase(t, "other-hash", "not-base64")
	if _, err := DecodeKGMWithDatabase(bytes.NewReader(syntheticKGMV5("wanted-hash")), "input.kgm", dbPath); !errors.Is(err, ErrKGGDatabaseEntryNotFound) {
		t.Fatalf("missing entry error=%v", err)
	}
	dbPath = makeSyntheticKGGDatabase(t, "wanted-hash", "not-base64")
	if _, err := DecodeKGMWithDatabase(bytes.NewReader(syntheticKGMV5("wanted-hash")), "input.kgm", dbPath); errors.Is(err, ErrKGGDatabaseEntryNotFound) || err == nil {
		t.Fatalf("matching entry was not passed to decoder: %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(dbPath))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("unexpected database sidecar files: %+v", entries)
	}
}

func TestDecodeKGMRejectsMalformedInput(t *testing.T) {
	if _, err := DecodeKGM(bytes.NewReader([]byte("not kgm")), "input.bin"); err == nil {
		t.Fatal("malformed KGM was accepted")
	}
}

func TestDecodeKGMV3UsesCanonicalAdapter(t *testing.T) {
	stream, err := DecodeKGM(bytes.NewReader(syntheticKGMV3()), "input.with-unusual-name")
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4)
	if _, err := stream.Read(buf); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeNCMRejectsMalformedInput(t *testing.T) {
	if _, err := DecodeNCM(bytes.NewReader([]byte("CTENFDAM")), "input.ncm"); err == nil {
		t.Fatal("malformed NCM was accepted")
	}
}

func TestDecodeKWMRejectsMalformedInput(t *testing.T) {
	if _, err := DecodeKWM(bytes.NewReader([]byte("yeelion-kuwo-tme")), "input.kwm"); err == nil {
		t.Fatal("malformed KWM was accepted")
	}
}

func TestLegacyAdaptersRejectMalformedInput(t *testing.T) {
	if _, err := DecodeTM(bytes.NewReader([]byte("QQMU")), "input.tm2"); err == nil {
		t.Fatal("malformed TM was accepted")
	}
	if _, err := DecodeXiami(bytes.NewReader([]byte("ifmt")), "input.xm"); err == nil {
		t.Fatal("malformed Xiami was accepted")
	}
	if _, err := DecodeXimalaya(bytes.NewReader([]byte("invalid")), "input.x2m"); err == nil {
		t.Fatal("malformed Ximalaya was accepted")
	}
}

func TestDecodeQMCStaticVariant(t *testing.T) {
	plain := append([]byte("fLaC"), bytes.Repeat([]byte{0x21}, 60)...)
	cipher, err := canonicalqmc.NewQmcCipherDecoder(nil)
	if err != nil {
		t.Fatal(err)
	}
	cipher.Decrypt(plain, 0)
	data := append(plain, []byte{0xff, 0xff, 0xff, 0xff}...)
	stream, err := DecodeQMC(bytes.NewReader(data), "input.qmc0")
	if err != nil {
		t.Fatal(err)
	}
	prefix := make([]byte, 4)
	if _, err := stream.Read(prefix); err != nil {
		t.Fatal(err)
	}
	if string(prefix) != "fLaC" {
		t.Fatalf("prefix = %q", prefix)
	}
}

func TestCandidateSuffixesUsesUpstreamFilenameRegistrations(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     []string
	}{
		{name: "kgm", filename: "track.kgm", want: []string{".kgm"}},
		{name: "kgm flac compound", filename: "track.kgm.flac", want: []string{".flac", ".kgm.flac"}},
		{name: "vpr flac compound", filename: "track.vpr.flac", want: []string{".flac", ".vpr.flac"}},
		{name: "kwm", filename: "track.kwm", want: []string{".kwm"}},
		{name: "ncm", filename: "track.ncm", want: []string{".ncm"}},
		{name: "qmc", filename: "track.qmc0", want: []string{".qmc0"}},
		{name: "qmc mgg suffix", filename: "track.mgg0", want: []string{".mgg0"}},
		{name: "qmc mflac suffix", filename: "track.mflac0", want: []string{".mflac0"}},
		{name: "tm", filename: "track.tm2", want: []string{".tm2"}},
		{name: "xiami", filename: "track.xm", want: []string{".xm"}},
		{name: "ximalaya", filename: "track.x2m", want: []string{".x2m"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CandidateSuffixes(tt.filename); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("CandidateSuffixes(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestCandidateSuffixesDoesNotInventUnsupportedCompoundSuffixes(t *testing.T) {
	for _, filename := range []string{"track.mgg0.flac", "track.mflac0.flac"} {
		got := CandidateSuffixes(filename)
		for _, suffix := range got {
			if suffix == ".mgg0.flac" || suffix == ".mflac0.flac" {
				t.Fatalf("CandidateSuffixes(%q) invented unsupported compound suffix %q", filename, suffix)
			}
		}
	}
}

func TestCandidateSuffixesDoesNotValidateContent(t *testing.T) {
	// A matching filename is only a candidate; this test intentionally supplies
	// no file and makes no claim about encrypted-format validity.
	if got := CandidateSuffixes("not-a-real-file.ncm"); !reflect.DeepEqual(got, []string{".ncm"}) {
		t.Fatalf("CandidateSuffixes returned %v, want filename-only candidate [.ncm]", got)
	}
}
