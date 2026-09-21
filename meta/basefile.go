package meta

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CSVFile identifies a CSV input and the model name derived from its filename.
type CSVFile struct {
	Model string
	Path  string
}

// CSVFiles returns CSV inputs in lexical order. It does not open the files, so
// callers retain clear ownership of every file descriptor they acquire.
func CSVFiles(path string) ([]CSVFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect %q: %w", path, err)
	}

	if !info.IsDir() {
		file, err := csvFile(path)
		if err != nil {
			return nil, err
		}
		return []CSVFile{file}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read directory %q: %w", path, err)
	}
	files := make([]CSVFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".csv") {
			continue
		}
		file, err := csvFile(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no CSV files found in %q", path)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func csvFile(path string) (CSVFile, error) {
	ext := filepath.Ext(path)
	if !strings.EqualFold(ext, ".csv") {
		return CSVFile{}, fmt.Errorf("%q is not a CSV file", path)
	}
	name := strings.TrimSuffix(filepath.Base(path), ext)
	if name == "" {
		return CSVFile{}, fmt.Errorf("cannot derive model name from %q", path)
	}
	return CSVFile{Model: name, Path: filepath.Clean(path)}, nil
}
