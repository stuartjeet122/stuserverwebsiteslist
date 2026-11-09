# Stuserver

A lightweight, server-rendered public directory for websites and APIs hosted on your personal server. The Go application only reads the JSON registry; the separate Cockpit plugin remains responsible for creating and maintaining its contents.

## Requirements

- Go 1.22 or newer to build
- No Node.js runtime or third-party Go modules

## Run locally

```sh
cp data/websites.example.json data/websites.json
WEBSITES_REGISTRY_PATH=./data/websites.json go run ./cmd/server
```

Open `http://127.0.0.1:8080`. `HOST`, `PORT`, `WEBSITES_REGISTRY_PATH` and `STATIC_ASSETS_PATH` can be set through the environment. Defaults are `127.0.0.1`, `8080`, `/home/stuserver/data/websites.json` and `static`.

Build a standalone executable:

```sh
go build -o stuserver ./cmd/server
./stuserver
```

## Registry

Use [data/websites.example.json](data/websites.example.json) as the schema reference. The application reads the registry at startup and checks its file metadata on requests; when the modification time or size changes, it reloads and validates the JSON. If the file is temporarily missing or invalid, the last valid registry remains in use. If no valid file has been read yet, an empty directory is rendered.

The public application never writes to the registry and has no management API. Only visible entries of supported types are rendered. HTTP(S) links and local static asset paths are validated before output.

Place Cockpit-managed preview images in `static/previews/` and icons in `static/icons/` (or the configured `STATIC_ASSETS_PATH`). These files are served read-only at runtime, so adding a preview does not require rebuilding the executable. Embedded CSS, JavaScript and default icons remain available as fallbacks. Registry asset paths are relative to the static directory, for example `/previews/my-site.webp` and `/icons/my-site.svg`. Supported preview formats are WebP, PNG and JPEG, plus SVG placeholders. The optional `preview_url` is accepted as metadata but never fetched or embedded.

## Deployment

1. Build on a compatible Linux system with `go build -o stuserver ./cmd/server`.
2. Install the executable under `/opt/stuserver/stuserver` and create `/opt/stuserver/static/previews` and `/opt/stuserver/static/icons` for Cockpit-managed assets. CSS, JavaScript and default icons are embedded in the executable.
3. Create a dedicated `stuserver` user and group. Grant it read access to the registry and static assets, with no write access.
4. Install `deploy/stuserver.service` to `/etc/systemd/system/stuserver.service`. Adjust paths and permissions if your registry is elsewhere.
5. Configure environment variables in `/etc/stuserver/stuserver.env` (example: [.env.example](.env.example)).
6. Enable the service with `systemctl enable --now stuserver`.
7. Adapt [deploy/nginx-stuserver.conf](deploy/nginx-stuserver.conf) for your Nginx setup. TLS and Cloudflare Tunnel remain managed by your existing infrastructure.

The example service binds to localhost by default and has systemd hardening enabled. `ProtectHome=false` is required for the default registry under `/home`; Unix ownership and permissions should grant the service account read-only access to the data directory and registry file.

Routes:

- `/` — public directory
- `/health` — minimal JSON health status
- `/robots.txt` — crawler policy
- `/favicon.ico` — redirects to the embedded SVG favicon
- `/static/*` — embedded CSS, JavaScript and default assets, plus read-only runtime preview images and icons

## Tests and checks

```sh
gofmt -w .
go vet ./...
go test ./...
go build -o stuserver ./cmd/server
```
