package handlers

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"stuserver"
	"stuserver/internal/registry"
)

func TestSafeExternalURL(t *testing.T) {
	for _, unsafe := range []string{"javascript:alert(1)", "data:text/html,test", "file:///etc/passwd", "//example.com"} {
		if got := safeExternalURL(unsafe); got != "" {
			t.Errorf("safeExternalURL(%q) = %q; want empty", unsafe, got)
		}
	}
	for _, safe := range []string{"https://example.com/a", "http://example.com"} {
		if got := safeExternalURL(safe); got == "" {
			t.Errorf("safeExternalURL(%q) unexpectedly rejected", safe)
		}
	}
}

func TestSanitizeAssetFilesUsesFallbackForMissingFiles(t *testing.T) {
	assets := fstest.MapFS{
		"previews/known.webp": &fstest.MapFile{Data: []byte("image")},
	}
	entries := sanitizeAssetFiles(sanitizeEntries([]registry.Entry{
		{Type: "website", PreviewImage: "/previews/known.webp", Icon: "/icons/missing.svg"},
	}), assets)
	if len(entries) != 1 || entries[0].PreviewImage != "/static/previews/known.webp" || entries[0].Icon != "" {
		t.Fatalf("unexpected asset sanitization: %#v", entries)
	}
	if _, err := fs.Stat(assets, "previews/known.webp"); err != nil {
		t.Fatal(err)
	}
}

func TestSafeAssetPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/previews/site.webp", "/static/previews/site.webp"},
		{"/icons/site.svg", "/static/icons/site.svg"},
		{"/previews/../../secret.webp", ""},
		{"https://example.com/image.jpg", ""},
		{"/previews/site.gif", ""},
	}
	for _, test := range tests {
		if got := safeAssetPath(test.input); got != test.want {
			t.Errorf("safeAssetPath(%q) = %q; want %q", test.input, got, test.want)
		}
	}
}

func TestPageDataCountsOnlyVisibleRegistryEntries(t *testing.T) {
	page := buildPageData(registry.Registry{
		Websites: []registry.Entry{{Type: "website", Name: "Site", Visible: true, Featured: true, Category: "Apps"}},
		APIs:     []registry.Entry{{Type: "api", Name: "API", Visible: true, Category: "Apps"}},
	})
	if page.TotalCount != 2 || page.WebsiteCount != 1 || page.APICount != 1 || len(page.Featured) != 1 {
		t.Fatalf("unexpected page counts: %#v", page)
	}
	if len(page.Categories) != 1 || page.Categories[0] != "Apps" {
		t.Fatalf("unexpected categories: %#v", page.Categories)
	}
}

func TestHomepageEscapesRegistryTextAndHealthRoute(t *testing.T) {
	path := filepath.Join(t.TempDir(), "websites.json")
	content := `{"websites":[{"id":"test","type":"website","name":"<script>alert(1)</script>","description":"safe","visible":true}]}`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	mux, err := NewMux(registry.NewCache(path, nil), stuserver.Files, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	homepage := httptest.NewRecorder()
	mux.ServeHTTP(homepage, httptest.NewRequest(http.MethodGet, "/", nil))
	if homepage.Code != http.StatusOK {
		t.Fatalf("homepage status = %d", homepage.Code)
	}
	if got := homepage.Body.String(); !strings.Contains(got, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("registry text was not HTML-escaped: %s", got)
	}
	if strings.Contains(homepage.Body.String(), "<script>alert(1)</script>") {
		t.Fatal("unescaped registry script was rendered")
	}

	health := httptest.NewRecorder()
	mux.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected health response: %d %q", health.Code, health.Body.String())
	}
}

func TestRuntimePreviewAssetsAreServedFromConfiguredDirectory(t *testing.T) {
	staticRoot := t.TempDir()
	previewDir := filepath.Join(staticRoot, "previews")
	if err := os.MkdirAll(previewDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(previewDir, "cockpit.webp"), []byte("preview data"), 0600); err != nil {
		t.Fatal(err)
	}
	registryPath := filepath.Join(t.TempDir(), "websites.json")
	content := `{"websites":[{"id":"site","type":"website","name":"Site","visible":true,"url":"https://example.com","preview_image":"/previews/cockpit.webp"}]}`
	if err := os.WriteFile(registryPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	mux, err := NewMux(registry.NewCache(registryPath, nil), stuserver.Files, staticRoot)
	if err != nil {
		t.Fatal(err)
	}

	homepage := httptest.NewRecorder()
	mux.ServeHTTP(homepage, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(homepage.Body.String(), `src="/static/previews/cockpit.webp"`) {
		t.Fatal("configured preview image was not rendered")
	}

	preview := httptest.NewRecorder()
	mux.ServeHTTP(preview, httptest.NewRequest(http.MethodGet, "/static/previews/cockpit.webp", nil))
	if preview.Code != http.StatusOK || preview.Body.String() != "preview data" {
		t.Fatalf("unexpected preview response: %d %q", preview.Code, preview.Body.String())
	}
}
