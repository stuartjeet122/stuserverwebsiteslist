package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

const maxRegistrySize = 4 << 20

func readFile(path string) (Registry, error) {
	file, err := os.Open(path)
	if err != nil {
		return Registry{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxRegistrySize+1))
	if err != nil {
		return Registry{}, fmt.Errorf("read registry: %w", err)
	}
	if len(data) > maxRegistrySize {
		return Registry{}, errors.New("registry exceeds maximum size")
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return Registry{}, fmt.Errorf("parse registry JSON: %w", err)
	}
	registry.Websites = supportedVisible(registry.Websites, "website")
	registry.APIs = supportedVisible(registry.APIs, "api")
	sortEntries(registry.Websites)
	sortEntries(registry.APIs)
	return registry, nil
}

func supportedVisible(entries []Entry, expectedType string) []Entry {
	result := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Type == expectedType && entry.Visible {
			result = append(result, entry)
		}
	}
	return result
}

func sortEntries(entries []Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Order < entries[j].Order
	})
}
