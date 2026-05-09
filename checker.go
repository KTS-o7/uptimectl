package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// CheckResult stores the outcome of a single health check.
type CheckResult struct {
	ServiceName string `json:"service_name"`
	Timestamp   string `json:"timestamp"` // RFC3339
	Success     bool   `json:"success"`
	StatusCode  int    `json:"status_code"`
	LatencyMs   int64  `json:"latency_ms"`
	Error       string `json:"error,omitempty"`
}

// CheckService performs a health check against the given service config.
func CheckService(svc ServiceConfig) CheckResult {
	result := CheckResult{
		ServiceName: svc.Name,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	client := &http.Client{
		Timeout: svc.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			// Follow redirects — many services redirect / to /some/path.
			return nil
		},
	}

	start := time.Now()
	req, err := http.NewRequest(svc.Method, svc.URL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("request creation failed: %v", err)
		result.Success = false
		result.LatencyMs = time.Since(start).Milliseconds()
		return result
	}

	req.Header.Set("User-Agent", "uptimectl/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	result.LatencyMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		result.Success = false
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode == svc.ExpectedStatus

	if !result.Success {
		result.Error = fmt.Sprintf("expected status %d, got %d", svc.ExpectedStatus, resp.StatusCode)
	}

	return result
}
