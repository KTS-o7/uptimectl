# uptimectl

Lightweight Go-based uptime monitor. Replaces Uptime Kuma with a single 6.5MB binary.

## Features

- HTTP health checks for multiple services
- JSON file storage (30 day history)
- Static dark-themed status page generation
- Systemd service for automatic monitoring
- ~2.5MB RAM at idle

## Setup

```bash
# Build
go build -ldflags="-s -w" -o uptimectl .

# Run as daemon
./uptimectl config.yaml
```

## Configuration

Edit `config.yaml` to add/remove services:

```yaml
services:
  - name: "My Service"
    url: "https://example.com/health"
    method: GET
    timeout: 10s
    expected_status: 200
```

## Systemd Service

```bash
cp uptimectl /opt/uptimectl/
cp config.yaml /opt/uptimectl/
systemctl enable --now uptimectl
```

## Status Page

The generated status page is a self-contained HTML file with:
- Real-time status indicators (green/red dots)
- 24h, 7d, and 30d uptime percentages
- Latency sparkline showing last 30 checks
- Dark theme, responsive design

## License

MIT
