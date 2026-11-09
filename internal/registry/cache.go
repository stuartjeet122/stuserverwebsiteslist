package registry

import (
	"log/slog"
	"os"
	"sync"
	"time"
)

type fileState struct {
	modTime time.Time
	size    int64
	exists  bool
}

func (s fileState) matches(other fileState) bool {
	return s.exists == other.exists && s.size == other.size && s.modTime.Equal(other.modTime)
}

type Cache struct {
	path    string
	logger  *slog.Logger
	mu      sync.Mutex
	data    Registry
	state   fileState
	checked bool
}

func NewCache(path string, logger *slog.Logger) *Cache {
	if logger == nil {
		logger = slog.Default()
	}
	return &Cache{
		path:   path,
		logger: logger,
		data:   emptyRegistry(),
	}
}

func (c *Cache) Get() Registry {
	c.mu.Lock()
	defer c.mu.Unlock()

	info, err := os.Stat(c.path)
	if err != nil {
		if !c.checked || c.state.exists {
			c.logger.Warn("registry unavailable; serving last valid data", "error", err)
		}
		c.state = fileState{exists: false}
		c.checked = true
		return clone(c.data)
	}

	current := fileState{modTime: info.ModTime(), size: info.Size(), exists: true}
	if c.checked && current.matches(c.state) {
		return clone(c.data)
	}

	data, err := readFile(c.path)
	if err != nil {
		c.logger.Error("registry reload failed; serving last valid data", "error", err)
		c.state = current
		c.checked = true
		return clone(c.data)
	}

	c.data = data
	c.state = current
	c.checked = true
	c.logger.Info("registry loaded", "websites", len(data.Websites), "apis", len(data.APIs))
	return clone(c.data)
}

func clone(source Registry) Registry {
	result := Registry{
		Version:  source.Version,
		Websites: append([]Entry(nil), source.Websites...),
		APIs:     append([]Entry(nil), source.APIs...),
	}
	for i := range result.Websites {
		result.Websites[i].Tags = append([]string(nil), result.Websites[i].Tags...)
	}
	for i := range result.APIs {
		result.APIs[i].Tags = append([]string(nil), result.APIs[i].Tags...)
	}
	if result.Websites == nil {
		result.Websites = make([]Entry, 0)
	}
	if result.APIs == nil {
		result.APIs = make([]Entry, 0)
	}
	return result
}
