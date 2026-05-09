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

	// Count overall status
	totalUp := 0
	anyDown := false
	for _, svc := range services {
		results := store.GetResults(svc.Name)
		if len(results) > 0 && results[0].Success {
			totalUp++
		} else if len(results) > 0 && !results[0].Success {
			anyDown = true
		}
	}

	overallStatus := "All Systems Operational"
	overallClass := "operational"
	if anyDown {
		overallStatus = "Degraded Performance"
		overallClass = "degraded"
	}
	if totalUp == 0 && len(services) > 0 {
		overallStatus = "Major Outage"
		overallClass = "outage"
	}

	genPageHeader(&b, overallStatus, overallClass, len(services))

	for _, svc := range services {
		results := store.GetResults(svc.Name)
		uptime90d := store.GetUptime(svc.Name, 90*24*time.Hour)

		// Determine current status
		statusText := "Operational"
		statusClass := "operational"
		if len(results) > 0 {
			latest := results[0]
			if !latest.Success {
				statusText = "Degraded"
				statusClass = "degraded"
			}
		}

		// Generate uptime bar (showing last 48 checks = ~24h at 30min intervals)
		barHTML := genUptimeBar(results, 48)

		// Format uptime with 2 decimal places, match githubstatus.com style
		uptimeStr := formatUptime(uptime90d)

		fmt.Fprintf(&b, `
      <div class="component">
        <div class="component-row">
          <div class="component-name">
            <span class="status-indicator %s"></span>
            <span>%s</span>
          </div>
          <div class="component-status">
            <span class="badge %s">%s</span>
          </div>
          <div class="component-uptime">%s%% uptime</div>
          <div class="component-bar">%s</div>
        </div>
      </div>`,
			statusClass, svc.Name,
			statusClass, statusText,
			uptimeStr,
			barHTML,
		)
	}

	genPageFooter(&b)

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

func genPageHeader(b *strings.Builder, status string, statusClass string, numServices int) {
	now := time.Now().UTC()

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Status · shenthar.me</title>
<style>
  :root {
    --bg: #ffffff;
    --bg-secondary: #f6f8fa;
    --text: #24292f;
    --text-secondary: #656d76;
    --text-muted: #8b949e;
    --border: #d0d7de;
    --border-light: #e8ecef;
    --green: #1a7f37;
    --green-bg: #dafbe1;
    --green-bar: #1a7f37;
    --yellow: #9a6700;
    --yellow-bg: #fff8c5;
    --yellow-bar: #d4a72c;
    --red: #cf222e;
    --red-bg: #ffebe9;
    --red-bar: #cf222e;
    --header-bg: #24292f;
    --header-text: #ffffff;
    --radius: 6px;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  /* Navigation bar */
  .nav {
    width: 100%;
    background: var(--header-bg);
    padding: 12px 0;
  }
  .nav-inner {
    max-width: 960px;
    margin: 0 auto;
    padding: 0 24px;
    display: flex;
    align-items: center;
    gap: 24px;
  }
  .nav-logo {
    color: var(--header-text);
    font-size: 14px;
    font-weight: 600;
    text-decoration: none;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .nav-logo svg { fill: var(--header-text); }
  .nav-links {
    display: flex;
    gap: 20px;
    font-size: 13px;
  }
  .nav-links a {
    color: rgba(255,255,255,0.75);
    text-decoration: none;
  }
  .nav-links a:hover { color: var(--header-text); }

  /* Main container */
  .container { max-width: 960px; width: 100%; padding: 0 24px; }

  /* Status banner */
  .status-banner {
    margin: 40px 0 32px;
    padding: 32px;
    border-radius: var(--radius);
    border: 1px solid var(--border-light);
  }
  .status-banner.operational {
    background: var(--green-bg);
    border-color: var(--green-bg);
  }
  .status-banner.degraded {
    background: var(--yellow-bg);
    border-color: var(--yellow-bg);
  }
  .status-banner.outage {
    background: var(--red-bg);
    border-color: var(--red-bg);
  }
  .status-icon {
    font-size: 32px;
    margin-bottom: 8px;
  }
  .status-banner h1 {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 4px;
  }
  .status-banner.operational h1 { color: var(--green); }
  .status-banner.degraded h1 { color: var(--yellow); }
  .status-banner.outage h1 { color: var(--red); }
  .status-banner p {
    font-size: 14px;
    color: var(--text-secondary);
  }

  /* Components table */
  .components-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid var(--border);
    font-size: 12px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .component {
    border-bottom: 1px solid var(--border-light);
  }
  .component-row {
    display: flex;
    align-items: center;
    padding: 16px 0;
    gap: 16px;
  }
  .component-name {
    flex: 0 0 220px;
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
    font-weight: 500;
  }
  .status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .status-indicator.operational { background: var(--green); }
  .status-indicator.degraded { background: var(--yellow); }
  .status-indicator.outage { background: var(--red); }
  .component-status {
    flex: 0 0 120px;
  }
  .badge {
    font-size: 12px;
    font-weight: 500;
    padding: 3px 10px;
    border-radius: 20px;
  }
  .badge.operational { background: var(--green-bg); color: var(--green); }
  .badge.degraded { background: var(--yellow-bg); color: var(--yellow); }
  .badge.outage { background: var(--red-bg); color: var(--red); }
  .component-uptime {
    flex: 0 0 100px;
    font-size: 13px;
    color: var(--text-secondary);
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  /* Uptime bar */
  .component-bar {
    flex: 1;
    min-width: 100px;
  }
  .uptime-bar {
    display: flex;
    gap: 2px;
    height: 16px;
    align-items: stretch;
  }
  .uptime-segment {
    flex: 1;
    border-radius: 2px;
    min-height: 6px;
  }
  .uptime-segment.up { background: var(--green-bar); opacity: 0.8; }
  .uptime-segment.down { background: var(--red-bar); opacity: 0.8; }
  .uptime-segment.unknown { background: var(--border); opacity: 0.4; }

  /* Footer */
  footer {
    margin-top: 48px;
    margin-bottom: 40px;
    padding-top: 24px;
    border-top: 1px solid var(--border-light);
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    color: var(--text-muted);
  }
  footer a { color: var(--text-secondary); text-decoration: none; }
  footer a:hover { color: var(--text); }

  @media (max-width: 768px) {
    .component-row { flex-wrap: wrap; gap: 8px; }
    .component-name { flex: 0 0 100%; }
    .component-status { flex: 0 0 auto; }
    .component-uptime { flex: 0 0 auto; }
    .component-bar { flex: 1 1 100%; }
    .status-banner { padding: 24px; }
    .status-banner h1 { font-size: 20px; }
    .nav-inner { padding: 0 16px; }
  }
</style>
</head>
<body>
<div class="nav">
  <div class="nav-inner">
    <a class="nav-logo" href="https://shenthar.me">
      <svg height="20" viewBox="0 0 16 16" width="20"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>
      <span>Status</span>
    </a>
    <div class="nav-links">
      <a href="https://shenthar.me">Home</a>
      <a href="https://github.com/KTS-o7/uptimectl">Source</a>
    </div>
  </div>
</div>
<div class="container">`)

	fmt.Fprintf(b, `  <div class="status-banner %s">
    <div class="status-icon">`, statusClass)
	if statusClass == "operational" {
		b.WriteString("&#10003;")
	} else if statusClass == "degraded" {
		b.WriteString("&#9888;")
	} else {
		b.WriteString("&#10007;")
	}
	fmt.Fprintf(b, `</div>
    <h1>%s</h1>
    <p>Monitoring %d services · Last checked %s</p>
  </div>
  <div class="components-header">
    <span>Components</span>
    <span>Status</span>
    <span>Uptime</span>
    <span>Recent checks</span>
  </div>`,
		status, numServices, now.Format("Jan 2, 2006 15:04 UTC"),
	)
}

func genPageFooter(b *strings.Builder) {
	now := time.Now().UTC().Format("Jan 2, 2006 15:04 UTC")
	fmt.Fprintf(b, `
  <footer>
    <span>Status · shenthar.me</span>
    <span>Updated %s · <a href="https://github.com/KTS-o7/uptimectl">uptimectl</a></span>
  </footer>
</div>
</body>
</html>`, now)
}

// genUptimeBar generates a horizontal bar showing green/red segments for recent checks.
func genUptimeBar(results []CheckResult, maxSegments int) string {
	if len(results) == 0 {
		return `<div class="uptime-bar"><div class="uptime-segment unknown" style="flex:1"></div></div>`
	}

	n := maxSegments
	if len(results) < n {
		n = len(results)
	}

	var bars strings.Builder
	bars.WriteString(`<div class="uptime-bar">`)

	for i := n - 1; i >= 0; i-- {
		r := results[i]
		if r.Success {
			bars.WriteString(`<div class="uptime-segment up"></div>`)
		} else {
			bars.WriteString(`<div class="uptime-segment down"></div>`)
		}
	}

	bars.WriteString(`</div>`)
	return bars.String()
}

func formatUptime(uptime float64) string {
	if uptime >= 99.995 {
		return "100.00"
	}
	return fmt.Sprintf("%.2f", uptime)
}
