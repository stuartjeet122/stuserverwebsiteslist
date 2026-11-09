package registry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadFileFiltersAndSortsEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "websites.json")
	content := `{
		"version": 1,
		"websites": [
			{"id":"later","type":"website","name":"Later","visible":true,"order":2},
			{"id":"hidden","type":"website","visible":false,"order":0},
			{"id":"unsupported","type":"other","visible":true,"order":1},
			{"id":"first","type":"website","name":"First","visible":true,"order":1}
		],
		"apis": [{"id":"api","type":"api","visible":true,"order":1}]
	}`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := readFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Websites) != 2 || got.Websites[0].ID != "first" || got.Websites[1].ID != "later" {
		t.Fatalf("unexpected websites: %#v", got.Websites)
	}
	if len(got.APIs) != 1 || got.APIs[0].ID != "api" {
		t.Fatalf("unexpected APIs: %#v", got.APIs)
	}
}

func TestReadFileRejectsMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "websites.json")
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readFile(path); err == nil {
		t.Fatal("expected malformed JSON to fail")
	}
}

func TestReadFileAcceptsEmptyRegistry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "websites.json")
	if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := readFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Websites) != 0 || len(got.APIs) != 0 {
		t.Fatalf("expected an empty registry, got %#v", got)
	}
}

func TestCacheRetainsLastValidRegistry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "websites.json")
	logger := newTestLogger(t)
	cache := NewCache(path, logger)
	valid := `{"websites":[{"id":"one","type":"website","visible":true}]}`
	if err := os.WriteFile(path, []byte(valid), 0600); err != nil {
		t.Fatal(err)
	}
	if got := cache.Get(); len(got.Websites) != 1 {
		t.Fatalf("initial load failed: %#v", got)
	}

	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, time.Now().Add(time.Second), time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if got := cache.Get(); len(got.Websites) != 1 || got.Websites[0].ID != "one" {
		t.Fatalf("cache did not retain last valid version: %#v", got)
	}
}

func TestCacheReloadsChangedRegistry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "websites.json")
	cache := NewCache(path, newTestLogger(t))
	first := `{"websites":[{"id":"first","type":"website","visible":true}]}`
	if err := os.WriteFile(path, []byte(first), 0600); err != nil {
		t.Fatal(err)
	}
	if got := cache.Get(); len(got.Websites) != 1 || got.Websites[0].ID != "first" {
		t.Fatalf("initial load failed: %#v", got)
	}

	second := `{"websites":[{"id":"second","type":"website","visible":true}]}`
	if err := os.WriteFile(path, []byte(second), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, time.Now().Add(time.Second), time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if got := cache.Get(); len(got.Websites) != 1 || got.Websites[0].ID != "second" {
		t.Fatalf("registry change was not reloaded: %#v", got)
	}
}
