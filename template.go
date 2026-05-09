package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateStatusPage writes the status HTML to the output path.
func GenerateStatusPage(store *ResultsStore, services []ServiceConfig, outputPath string) error {
	var b strings.Builder

	genPageHeader(&b)

	for _, svc := range services {
		results := store.GetResults(svc.Name)
		uptime24h := store.GetUptime(svc.Name, 24*time.Hour)
		uptime7d := store.GetUptime(svc.Name, 7*24*time.Hour)
		uptime30d := store.GetUptime(svc.Name, 30*24*time.Hour)

		// Determine current status
		status := "up"
		statusClass := "up"
		if len(results) > 0 {
			latest := results[0]
			if !latest.Success {
				status = "down"
				statusClass = "down"
				if latest.Error != "" {
					status += ": " + latest.Error
				}
			}
		}

		// Convert timestamp to relative
		lastCheckRel := "never"
		if len(results) > 0 {
			t, err := time.Parse(time.RFC3339, results[0].Timestamp)
			if err == nil {
				lastCheckRel = timeAgo(t)
			}
		}

		// History sparkline: last 30 checks
		sparkline := genSparkline(results, 30)

		fmt.Fprintf(&b, `
    <div class="service-card">
      <div class="service-header">
        <div class="service-name">
          <span class="status-dot %s"></span>
          <span>%s</span>
        </div>
        <div class="status-badge %s">%s</div>
      </div>
      <div class="service-details">
        <div class="stat">
          <span class="stat-label">24h</span>
          <span class="stat-value">%.1f%%</span>
        </div>
        <div class="stat">
          <span class="stat-label">7d</span>
          <span class="stat-value">%.1f%%</span>
        </div>
        <div class="stat">
          <span class="stat-label">30d</span>
          <span class="stat-value">%.1f%%</span>
        </div>
        <div class="stat">
          <span class="stat-label">Checks</span>
          <span class="stat-value">%d</span>
        </div>
      </div>
      <div class="service-footer">
        <div class="last-check">Checked %s</div>
        <div class="sparkline">%s</div>
      </div>
    </div>`,
			statusClass, svc.Name,
			statusClass, status,
			uptime24h, uptime7d, uptime30d,
			len(results),
			lastCheckRel,
			sparkline,
		)
	}

	genPageFooter(&b, len(services))

	// Write to output
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	tmpPath := outputPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("writing temp HTML: %w", err)
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		return fmt.Errorf("renaming temp HTML: %w", err)
	}

	return nil
}

func genPageHeader(b *strings.Builder) {
	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Status · shenthar.me</title>
<style>
  :root {
    --bg: #0a0a0f;
    --card-bg: #111118;
    --border: #1e1e2a;
    --text: #c8c8d4;
    --text-muted: #6b6b7d;
    --accent: #6c5ce7;
    --green: #00d68f;
    --red: #ff6b6b;
    --radius: 12px;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 40px 16px;
  }
  .container { max-width: 900px; width: 100%; }
  header {
    text-align: center;
    margin-bottom: 40px;
  }
  header h1 {
    font-size: 28px;
    font-weight: 700;
    margin-bottom: 8px;
    letter-spacing: -0.5px;
  }
  header h1 span { color: var(--accent); }
  header p {
    color: var(--text-muted);
    font-size: 14px;
  }
  .grid {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .service-card {
    background: var(--card-bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 20px 24px;
    transition: border-color 0.2s;
  }
  .service-card:hover { border-color: #2a2a3a; }
  .service-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }
  .service-name {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 16px;
    font-weight: 600;
  }
  .status-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .status-dot.up { background: var(--green); box-shadow: 0 0 8px rgba(0,214,143,0.4); }
  .status-dot.down { background: var(--red); box-shadow: 0 0 8px rgba(255,107,107,0.4); }
  .status-badge {
    font-size: 12px;
    font-weight: 600;
    padding: 4px 12px;
    border-radius: 20px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .status-badge.up { background: rgba(0,214,143,0.12); color: var(--green); }
  .status-badge.down { background: rgba(255,107,107,0.12); color: var(--red); }
  .service-details {
    display: flex;
    gap: 24px;
    margin-bottom: 12px;
  }
  .stat {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .stat-label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .stat-value {
    font-size: 18px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }
  .service-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .last-check {
    font-size: 12px;
    color: var(--text-muted);
  }
  .sparkline {
    display: flex;
    gap: 2px;
    align-items: flex-end;
    height: 20px;
  }
  .spark-bar {
    width: 6px;
    border-radius: 2px 2px 0 0;
    min-height: 3px;
    transition: opacity 0.2s;
  }
  .spark-bar.up { background: var(--green); opacity: 0.7; }
  .spark-bar.down { background: var(--red); opacity: 0.7; }
  .spark-bar:hover { opacity: 1; }
  footer {
    margin-top: 48px;
    text-align: center;
    font-size: 12px;
    color: var(--text-muted);
  }
  footer a { color: var(--accent); text-decoration: none; }
  @media (max-width: 600px) {
    body { padding: 20px 12px; }
    .service-details { gap: 16px; flex-wrap: wrap; }
    .stat-value { font-size: 16px; }
    .service-card { padding: 16px; }
  }
</style>
</head>
<body>
<div class="container">
<header>
  <h1><span>Status</span> · shenthar.me</h1>
  <p>Uptime monitoring for all services</p>
</header>
<div class="grid">
`)
}

func genPageFooter(b *strings.Builder, numServices int) {
	now := time.Now().UTC().Format("Mon Jan 02 15:04 UTC 2006")
	fmt.Fprintf(b, `
</div>
<footer>
  <p>Monitoring %d services · Last generated %s</p>
  <p><a href="https://github.com/KTS-o7/uptimectl">uptimectl</a></p>
</footer>
</div>
</body>
</html>`, numServices, now)
}

// genSparkline generates a sparkline div with colored bars for recent results.
func genSparkline(results []CheckResult, maxBars int) string {
	if len(results) == 0 {
		return `<span style="color:var(--text-muted);font-size:11px;">no data</span>`
	}

	n := maxBars
	if len(results) < n {
		n = len(results)
	}

	var bars strings.Builder
	for i := n - 1; i >= 0; i-- {
		r := results[i]
		cls := "up"
		if !r.Success {
			cls = "down"
		}
		// Scale height based on latency (min 3px, max 18px)
		h := 3 + (r.LatencyMs * 15 / 2000)
		if h > 18 {
			h = 18
		}
		if h < 3 {
			h = 3
		}
		fmt.Fprintf(&bars, `<div class="spark-bar %s" style="height:%dpx" title="%s"></div>`, cls, h, r.Timestamp)
	}

	return bars.String()
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", m)
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
