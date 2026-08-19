package externalmaterial

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const MaxEKeyFileSize = 16 << 10

// ReadEKeyFile reads the canonical upstream Base64 EKey representation.
// Only a UTF-8 BOM and surrounding whitespace are ignored.
func ReadEKeyFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("external QMC key file could not be opened")
	}
	defer f.Close()
	data, err := ioReadBounded(f, MaxEKeyFileSize)
	if err != nil {
		return nil, err
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, errors.New("external QMC key file is empty")
	}
	if _, err := base64.StdEncoding.DecodeString(text); err != nil {
		return nil, errors.New("external QMC key file is not valid Base64")
	}
	return []byte(text), nil
}

func ioReadBounded(f *os.File, max int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, errors.New("external QMC key file could not be read")
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("external QMC key file exceeds %d bytes", max)
	}
	return data, nil
}
