package containermeta

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"testing"

	"musicunlocker/internal/audiometa"
	"musicunlocker/internal/naming"
	"musicunlocker/internal/probe"
)

var testNCMKeyCore = []byte{0x68, 0x7a, 0x48, 0x52, 0x41, 0x6d, 0x73, 0x6f, 0x35, 0x6b, 0x49, 0x6e, 0x62, 0x61, 0x78, 0x57}
var testNCMKeyMeta = []byte{0x23, 0x31, 0x34, 0x6c, 0x6a, 0x6b, 0x5f, 0x21, 0x5c, 0x5d, 0x26, 0x30, 0x55, 0x3c, 0x27, 0x28}

func ecbEncrypt(key, plain []byte) []byte {
	b, _ := aes.NewCipher(key)
	out := make([]byte, len(plain))
	for i := 0; i < len(plain); i += b.BlockSize() {
		b.Encrypt(out[i:i+b.BlockSize()], plain[i:i+b.BlockSize()])
	}
	return out
}

func pkcs7(data []byte) []byte {
	n := aes.BlockSize - len(data)%aes.BlockSize
	return append(append([]byte{}, data...), bytes.Repeat([]byte{byte(n)}, n)...)
}

func syntheticNCM(t *testing.T, meta any) []byte {
	t.Helper()
	jsonMeta, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	keyPlain := append(bytes.Repeat([]byte{0}, 17), []byte("audio-key")...)
	keyRaw := ecbEncrypt(testNCMKeyCore, pkcs7(keyPlain))
	for i := range keyRaw {
		keyRaw[i] ^= 0x64
	}
	metaPlain := append([]byte("music:"), jsonMeta...)
	metaRaw := ecbEncrypt(testNCMKeyMeta, pkcs7(metaPlain))
	encoded := []byte(base64.StdEncoding.EncodeToString(metaRaw))
	for i := range encoded {
		encoded[i] ^= 0x63
	}
	prefix := []byte("163 key(Don't modify):")
	encoded = append(prefix, encoded...)
	cover := []byte{0}
	b := append([]byte("CTENFDAM"), 0, 0)
	var u [4]byte
	binary.LittleEndian.PutUint32(u[:], uint32(len(keyRaw)))
	b = append(b, u[:]...)
	b = append(b, keyRaw...)
	binary.LittleEndian.PutUint32(u[:], uint32(len(encoded)))
	b = append(b, u[:]...)
	b = append(b, encoded...)
	b = append(b, 0, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(u[:], uint32(len(cover)))
	b = append(b, u[:]...)
	binary.LittleEndian.PutUint32(u[:], uint32(len(cover)))
	b = append(b, u[:]...)
	return append(b, cover...)
}

func TestReadNCMMetadataOfflineAndDeterministic(t *testing.T) {
	data := syntheticNCM(t, map[string]any{
		"musicName": "Salvation", "artist": [][]any{{"Gabrielle Aplin", 1}, {"Guest Artist", 2}}, "album": "English Rain",
	})
	meta, err := Read(context.Background(), probe.FamilyNetease, bytes.NewReader(data), "synthetic.ncm")
	if err != nil || meta.Title != "Salvation" || meta.Artist != "Gabrielle Aplin / Guest Artist" || meta.Album != "English Rain" {
		t.Fatalf("metadata=%#v err=%v", meta, err)
	}
}

func TestReadNCMMetadataMalformedUnavailable(t *testing.T) {
	meta, err := Read(context.Background(), probe.FamilyNetease, bytes.NewReader([]byte("CTENFDAM")), "bad.ncm")
	if err == nil || meta.Available() {
		t.Fatalf("metadata=%#v err=%v", meta, err)
	}
}

func TestReadUnsupportedFamilyUnavailable(t *testing.T) {
	if _, err := Read(context.Background(), probe.FamilyQMC, bytes.NewReader(nil), "x"); err == nil {
		t.Fatal("unsupported family unexpectedly returned metadata")
	}
}

func TestContainerMetadataFeedsExistingNamingModes(t *testing.T) {
	container := audiometa.Metadata{Title: "Salvation", Artist: "Gabrielle Aplin", Album: "English Rain"}
	merged := audiometa.Merge(audiometa.Metadata{}, container)
	for _, tc := range []struct {
		mode naming.Mode
		want string
	}{{naming.Title, "Salvation"}, {naming.ArtistTitle, "Gabrielle Aplin - Salvation"}, {naming.TitleArtist, "Salvation - Gabrielle Aplin"}} {
		got, err := naming.Resolve(naming.Config{Mode: tc.mode}, "source", merged, merged.Available())
		if err != nil || got.Stem != tc.want {
			t.Fatalf("mode %s: %#v err=%v", tc.mode, got, err)
		}
	}
	for template, want := range map[string]string{"{title}-{artist}": "Salvation-Gabrielle Aplin", "{album} - {title}": "English Rain - Salvation"} {
		got, err := naming.Resolve(naming.Config{Mode: naming.Custom, CustomTemplate: template}, "source", merged, true)
		if err != nil || got.Stem != want {
			t.Fatalf("template %s: %#v err=%v", template, got, err)
		}
	}
}
