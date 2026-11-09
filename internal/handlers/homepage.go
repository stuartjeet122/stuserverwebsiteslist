package handlers

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"stuserver/internal/registry"
)

type PageData struct {
	Websites     []registry.Entry
	APIs         []registry.Entry
	Featured     []registry.Entry
	Categories   []string
	WebsiteCount int
	APICount     int
	TotalCount   int
	Year         int
}

type Handler struct {
	registry    *registry.Cache
	template    *template.Template
	static      http.Handler
	staticFiles fs.FS
}

func NewMux(cache *registry.Cache, assets embed.FS, dynamicAssetsPath string) (http.Handler, error) {
	embeddedStatic, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, fmt.Errorf("open embedded static files: %w", err)
	}
	staticFS := assetOverlay{
		embedded: embeddedStatic,
		dynamic:  os.DirFS(dynamicAssetsPath),
	}
	templates, err := template.New("layout.html").Funcs(template.FuncMap{
		"assetPath": safeAssetPath,
		"safeURL":   safeExternalURL,
	}).ParseFS(assets, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse HTML templates: %w", err)
	}

	handler := &Handler{
		registry:    cache,
		template:    templates,
		staticFiles: staticFS,
		static: http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=3600")
			http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
		})),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.homepage)
	mux.HandleFunc("/health", handler.health)
	mux.HandleFunc("/robots.txt", handler.robots)
	mux.HandleFunc("/favicon.ico", handler.favicon)
	mux.Handle("/static/", handler.static)
	return securityHeaders(mux), nil
}

func (h *Handler) homepage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := buildPageData(h.registry.Get())
	data.Websites = sanitizeAssetFiles(data.Websites, h.staticFiles)
	data.APIs = sanitizeAssetFiles(data.APIs, h.staticFiles)
	data.Featured = make([]registry.Entry, 0)
	for _, entry := range append(append([]registry.Entry(nil), data.Websites...), data.APIs...) {
		if entry.Featured {
			data.Featured = append(data.Featured, entry)
		}
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.ExecuteTemplate(w, "layout.html", data); err != nil {
		slog.Error("render homepage", "error", err)
	}
}

func buildPageData(source registry.Registry) PageData {
	websites := sanitizeEntries(source.Websites)
	apis := sanitizeEntries(source.APIs)
	featured := make([]registry.Entry, 0)
	categorySet := make(map[string]struct{})
	for _, group := range [][]registry.Entry{websites, apis} {
		for _, entry := range group {
			if entry.Featured {
				featured = append(featured, entry)
			}
			if strings.TrimSpace(entry.Category) != "" {
				categorySet[entry.Category] = struct{}{}
			}
		}
	}
	categories := make([]string, 0, len(categorySet))
	for category := range categorySet {
		categories = append(categories, category)
	}
	sortStrings(categories)

	return PageData{
		Websites: websites, APIs: apis, Featured: featured, Categories: categories,
		WebsiteCount: len(websites), APICount: len(apis),
		TotalCount: len(websites) + len(apis), Year: time.Now().Year(),
	}
}

func sanitizeEntries(entries []registry.Entry) []registry.Entry {
	result := make([]registry.Entry, 0, len(entries))
	for _, entry := range entries {
		entry.URL = safeExternalURL(entry.URL)
		entry.BaseURL = safeExternalURL(entry.BaseURL)
		entry.SwaggerURL = safeExternalURL(entry.SwaggerURL)
		entry.OpenAPIURL = safeExternalURL(entry.OpenAPIURL)
		entry.PreviewURL = safeExternalURL(entry.PreviewURL)
		entry.PreviewImage = safeAssetPath(entry.PreviewImage)
		entry.Icon = safeAssetPath(entry.Icon)
		result = append(result, entry)
	}
	return result
}

func sanitizeAssetFiles(entries []registry.Entry, assets fs.FS) []registry.Entry {
	for i := range entries {
		for _, field := range []*string{&entries[i].PreviewImage, &entries[i].Icon} {
			if *field == "" {
				continue
			}
			assetPath := strings.TrimPrefix(*field, "/static/")
			if _, err := fs.Stat(assets, assetPath); err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					slog.Warn("static asset could not be checked", "error", err)
				}
				*field = ""
			}
		}
	}
	return entries
}

type assetOverlay struct {
	embedded fs.FS
	dynamic  fs.FS
}

func (a assetOverlay) Open(name string) (fs.File, error) {
	if fs.ValidPath(name) && (strings.HasPrefix(name, "previews/") || strings.HasPrefix(name, "icons/")) {
		file, err := a.dynamic.Open(name)
		if err == nil {
			info, statErr := file.Stat()
			if statErr != nil {
				_ = file.Close()
				return nil, statErr
			}
			if info.IsDir() {
				_ = file.Close()
				return nil, fs.ErrNotExist
			}
			return file, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return a.embedded.Open(name)
}

func safeExternalURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil {
		return ""
	}
	return parsed.String()
}

func safeAssetPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return ""
	}
	assetPath := strings.TrimPrefix(parsed.Path, "/")
	assetPath = strings.TrimPrefix(assetPath, "static/")
	if strings.HasPrefix(assetPath, "previews/") {
		ext := strings.ToLower(path.Ext(assetPath))
		if ext != ".webp" && ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".svg" {
			return ""
		}
	} else if !strings.HasPrefix(assetPath, "icons/") {
		return ""
	}
	for _, segment := range strings.Split(assetPath, "/") {
		if segment == "." || segment == ".." {
			return ""
		}
	}
	return "/static/" + assetPath
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && strings.ToLower(values[j]) < strings.ToLower(values[j-1]); j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self'; style-src 'self'; script-src 'self'; frame-src 'none'; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
