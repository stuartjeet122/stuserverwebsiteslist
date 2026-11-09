package stuserver

import "embed"

//go:embed templates/*.html templates/partials/*.html static/css/* static/js/* static/icons/* static/previews/*
var Files embed.FS
