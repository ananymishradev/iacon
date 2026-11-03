package detect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func SaveAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	data, err := json.MarshalIndent(value, "", "  ")
	if err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func LoadState(path string) (*Incident, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var incident Incident
	if err := json.Unmarshal(data, &incident); err != nil {
		return nil, fmt.Errorf("decode incident state: %w", err)
	}
	return &incident, nil
}

func Rotate(dir string, keep int) error {
	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	sort.Strings(entries)
	if keep < 0 {
		keep = 0
	}
	for len(entries) > keep {
		if err := os.Remove(entries[0]); err != nil {
			return err
		}
		entries = entries[1:]
	}
	return nil
}
