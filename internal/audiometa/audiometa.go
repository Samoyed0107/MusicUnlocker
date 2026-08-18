// Package audiometa reads a deliberately small, offline subset of audio tags.
package audiometa

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

const maxTagBytes = 4 << 20

var ErrUnavailable = errors.New("audio metadata unavailable")

type Metadata struct {
	Title, Artist, Album, AlbumArtist string
	Track, TrackTotal                 int
	Disc, DiscTotal                   int
	Year                              int
}

func ReadFile(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxTagBytes))
	if err != nil {
		return Metadata{}, err
	}
	if len(b) >= 3 && bytes.Equal(b[:3], []byte("ID3")) {
		return parseID3(b)
	}
	if len(b) >= 4 && bytes.Equal(b[:4], []byte("fLaC")) {
		return parseFLAC(b)
	}
	if len(b) >= 4 && bytes.Equal(b[:4], []byte("OggS")) {
		return parseOgg(b)
	}
	if len(b) >= 8 && (bytes.Equal(b[4:8], []byte("ftyp")) || bytes.Contains(b[:min(len(b), 64)], []byte("ftyp"))) {
		return parseMP4(b)
	}
	return Metadata{}, ErrUnavailable
}

func parseID3(b []byte) (Metadata, error) {
	if len(b) < 10 {
		return Metadata{}, ErrUnavailable
	}
	sz, ok := syncsafe(b[6:10])
	if !ok || sz > maxTagBytes-10 || 10+sz > len(b) {
		return Metadata{}, ErrUnavailable
	}
	end := 10 + sz
	var m Metadata
	version := b[3]
	for p := 10; p+10 <= end; {
		id := string(b[p : p+4])
		n := int(binary.BigEndian.Uint32(b[p+4 : p+8]))
		if version == 4 {
			var sizeOK bool
			n, sizeOK = syncsafe(b[p+4 : p+8])
			if !sizeOK {
				return Metadata{}, ErrUnavailable
			}
		}
		p += 10
		if n < 0 || p+n > end {
			return Metadata{}, ErrUnavailable
		}
		if n == 0 || id[0] == 0 {
			break
		}
		value := decodeID3Text(b[p : p+n])
		switch id {
		case "TIT2":
			m.Title = value
		case "TPE1":
			m.Artist = value
		case "TALB":
			m.Album = value
		case "TPE2":
			m.AlbumArtist = value
		case "TRCK":
			m.Track, m.TrackTotal = pair(value)
		case "TPOS":
			m.Disc, m.DiscTotal = pair(value)
		case "TYER", "TDRC":
			m.Year = year(value)
		}
		p += n
	}
	if m == (Metadata{}) {
		return m, ErrUnavailable
	}
	return m, nil
}

func syncsafe(b []byte) (int, bool) {
	if len(b) != 4 || b[0]&0x80 != 0 || b[1]&0x80 != 0 || b[2]&0x80 != 0 || b[3]&0x80 != 0 {
		return 0, false
	}
	return int(b[0])<<21 | int(b[1])<<14 | int(b[2])<<7 | int(b[3]), true
}
func decodeID3Text(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	enc := b[0]
	data := bytes.TrimRight(b[1:], "\x00 ")
	if enc == 1 || enc == 2 {
		data = bytes.TrimPrefix(data, []byte{0xff, 0xfe})
		data = bytes.TrimPrefix(data, []byte{0xfe, 0xff})
	}
	return strings.TrimSpace(string(data))
}

func parseFLAC(b []byte) (Metadata, error) {
	p := 4
	var m Metadata
	for p+4 <= len(b) {
		last := b[p]&0x80 != 0
		typ := b[p] & 0x7f
		n := int(b[p+1])<<16 | int(b[p+2])<<8 | int(b[p+3])
		p += 4
		if n > maxTagBytes || p+n > len(b) {
			return Metadata{}, ErrUnavailable
		}
		if typ == 4 {
			m = parseVorbisComment(b[p : p+n])
		}
		p += n
		if last {
			break
		}
	}
	if m == (Metadata{}) {
		return m, ErrUnavailable
	}
	return m, nil
}

func parseOgg(b []byte) (Metadata, error) {
	var all []byte
	for p := 0; p+27 <= len(b); {
		if !bytes.Equal(b[p:p+4], []byte("OggS")) {
			break
		}
		n := int(b[p+26])
		if p+27+n > len(b) {
			break
		}
		seg := b[p+27 : p+27+n]
		q := p + 27 + n
		for _, size := range seg {
			if q+int(size) > len(b) {
				return Metadata{}, ErrUnavailable
			}
			all = append(all, b[q:q+int(size)]...)
			q += int(size)
		}
		p = q
	}
	if len(all) == 0 {
		return Metadata{}, ErrUnavailable
	}
	if i := bytes.Index(all, []byte{0x03, 'v', 'o', 'r', 'b', 'i', 's'}); i >= 0 {
		all = all[i+7:]
	} else if i := bytes.Index(all, []byte("OpusTags")); i >= 0 {
		all = all[i+8:]
	}
	m := parseVorbisComment(all)
	if m == (Metadata{}) {
		return m, ErrUnavailable
	}
	return m, nil
}

func parseVorbisComment(b []byte) Metadata {
	var m Metadata
	r := bytes.NewReader(b)
	var n uint32
	if binary.Read(r, binary.LittleEndian, &n) != nil || n > maxTagBytes {
		return m
	}
	if _, err := io.CopyN(io.Discard, r, int64(n)); err != nil {
		return m
	}
	if binary.Read(r, binary.LittleEndian, &n) != nil || n > 10000 {
		return m
	}
	for i := uint32(0); i < n; i++ {
		var l uint32
		if binary.Read(r, binary.LittleEndian, &l) != nil || l > maxTagBytes {
			return Metadata{}
		}
		x := make([]byte, l)
		if _, err := io.ReadFull(r, x); err != nil {
			return Metadata{}
		}
		kv := strings.SplitN(string(x), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, val := strings.ToLower(kv[0]), strings.TrimSpace(kv[1])
		switch key {
		case "title":
			m.Title = val
		case "artist":
			m.Artist = val
		case "album":
			m.Album = val
		case "albumartist":
			m.AlbumArtist = val
		case "tracknumber":
			m.Track, _ = strconv.Atoi(val)
		case "tracktotal":
			m.TrackTotal, _ = strconv.Atoi(val)
		case "discnumber":
			m.Disc, _ = strconv.Atoi(val)
		case "disctotal":
			m.DiscTotal, _ = strconv.Atoi(val)
		case "date", "year":
			m.Year = year(val)
		}
	}
	return m
}

func parseMP4(b []byte) (Metadata, error) {
	var m Metadata
	var walk func([]byte)
	walk = func(x []byte) {
		for p := 0; p+8 <= len(x); {
			n := int(binary.BigEndian.Uint32(x[p : p+4]))
			typ := string(x[p+4 : p+8])
			if n < 8 || n > len(x)-p {
				return
			}
			body := x[p+8 : p+n]
			if typ == "ilst" || typ == "moov" || typ == "udta" {
				walk(body)
			} else if typ == "meta" && len(body) >= 4 {
				walk(body[4:])
			} else if len(typ) == 4 && (typ[0] == '\xa9' || typ == "aART" || typ == "trkn" || typ == "disk") {
				if len(body) >= 16 && string(body[4:8]) == "data" {
					v := body[16:]
					switch typ {
					case "\xa9nam":
						m.Title = string(v)
					case "\xa9ART":
						m.Artist = string(v)
					case "\xa9alb":
						m.Album = string(v)
					case "aART":
						m.AlbumArtist = string(v)
					case "\xa9day":
						m.Year = year(string(v))
					}
				} else if (typ == "trkn" || typ == "disk") && len(body) >= 24 {
					if typ == "trkn" {
						m.Track = int(binary.BigEndian.Uint16(body[18:20]))
						m.TrackTotal = int(binary.BigEndian.Uint16(body[20:22]))
					} else {
						m.Disc = int(binary.BigEndian.Uint16(body[18:20]))
						m.DiscTotal = int(binary.BigEndian.Uint16(body[20:22]))
					}
				}
			}
			p += n
		}
	}
	walk(b)
	if m == (Metadata{}) {
		return m, ErrUnavailable
	}
	return m, nil
}

func pair(v string) (int, int) {
	parts := strings.SplitN(v, "/", 2)
	a, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	if len(parts) == 1 {
		return a, 0
	}
	c, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return a, c
}
func year(v string) int {
	v = strings.TrimSpace(v)
	if len(v) > 4 {
		v = v[:4]
	}
	n, _ := strconv.Atoi(v)
	return n
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
