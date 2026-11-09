package registry

type Registry struct {
	Version  int     `json:"version"`
	Websites []Entry `json:"websites"`
	APIs     []Entry `json:"apis"`
}

type Entry struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	URL          string   `json:"url,omitempty"`
	PreviewImage string   `json:"preview_image,omitempty"`
	PreviewURL   string   `json:"preview_url,omitempty"`
	BaseURL      string   `json:"base_url,omitempty"`
	SwaggerURL   string   `json:"swagger_url,omitempty"`
	OpenAPIURL   string   `json:"openapi_url,omitempty"`
	Version      string   `json:"version,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	Featured     bool     `json:"featured"`
	Visible      bool     `json:"visible"`
	Status       string   `json:"status,omitempty"`
	Order        int      `json:"order"`
}

func emptyRegistry() Registry {
	return Registry{
		Websites: make([]Entry, 0),
		APIs:     make([]Entry, 0),
	}
}
