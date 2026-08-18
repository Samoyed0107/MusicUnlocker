package app

import (
	"errors"
	"os"
	"path/filepath"
)

type SelectedInput struct {
	Path        string `json:"path"`
	DisplayName string `json:"displayName"`
	Kind        string `json:"kind"`
}

var ErrInvalidOpenPath = &ApplicationError{Code: "INVALID_OPEN_PATH", Detail: "the output path is not an accessible absolute path"}

type FolderOpener interface{ Open(string) error }

func SuggestOutputFolder(inputs []string) (string, error) {
	if len(inputs) != 1 || inputs[0] == "" {
		return "", nil
	}
	p, err := filepath.Abs(filepath.Clean(inputs[0]))
	if err != nil {
		return "", err
	}
	info, err := os.Stat(p)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return filepath.Join(filepath.Dir(p), filepath.Base(p)+" - Unlocked"), nil
	}
	return filepath.Join(filepath.Dir(p), "Unlocked"), nil
}

func DescribeInputs(inputs []string) []SelectedInput {
	items := make([]SelectedInput, 0, len(inputs))
	for _, input := range inputs {
		kind := "Path"
		if info, err := os.Stat(input); err == nil {
			if info.IsDir() {
				kind = "Folder"
			} else {
				kind = "File"
			}
		}
		items = append(items, SelectedInput{Path: input, DisplayName: filepath.Base(filepath.Clean(input)), Kind: kind})
	}
	return items
}

func validateOpenPath(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", ErrInvalidOpenPath
	}
	clean := filepath.Clean(path)
	info, err := os.Stat(clean)
	if err != nil {
		return "", ErrInvalidOpenPath
	}
	if info.IsDir() {
		return clean, nil
	}
	parent := filepath.Dir(clean)
	if parent == clean {
		return "", ErrInvalidOpenPath
	}
	if parentInfo, err := os.Stat(parent); err != nil || !parentInfo.IsDir() {
		return "", ErrInvalidOpenPath
	}
	return parent, nil
}

func openOutputFolder(path string, opener FolderOpener) error {
	clean, err := validateOpenPath(path)
	if err != nil {
		return err
	}
	if opener == nil {
		return errors.New("output folder opener unavailable")
	}
	return opener.Open(clean)
}
