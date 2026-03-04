package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	autostart "github.com/ProtonMail/go-autostart"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	surgeBaseURL = "http://127.0.0.1:1700"

	iconModeDock    = "dock"
	iconModeMenuBar = "menu_bar"
)

// App struct wraps the Surge engine via HTTP API
type App struct {
	ctx      context.Context
	token    string
	surgeBin string

	surgeProcess *exec.Cmd
	sseCancel    context.CancelFunc

	autostartApp *autostart.App
	iconMode     string
	shutdownOnce sync.Once
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// --------------------------------------------------------------------------
// Wails lifecycle
// --------------------------------------------------------------------------

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.initAutostart()
	a.loadIconMode()

	// 1. Resolve surge binary path (Finder/Dock launches may have limited PATH)
	a.surgeBin = a.resolveSurgeBinary()
	if a.surgeBin == "" {
		wailsRuntime.LogError(a.ctx, "surge binary not found (checked SURGE_BIN, common paths, and PATH)")
	} else {
		wailsRuntime.LogInfo(a.ctx, fmt.Sprintf("Using surge binary: %s", a.surgeBin))
	}

	// 2. Discover token
	a.token = a.discoverToken()

	// 3. Ensure server is reachable (either existing or spawned)
	serverReady := a.healthCheck()
	if !serverReady {
		if !a.startSurgeServer() {
			wailsRuntime.LogError(a.ctx, "surge server unavailable and failed to spawn")
			return
		}
		serverReady = a.waitForServer(10 * time.Second)
		if !serverReady {
			wailsRuntime.LogError(a.ctx, "surge server did not become ready in time")
			return
		}
	}

	// 4. Start SSE event stream -> Wails events
	sseCtx, cancel := context.WithCancel(context.Background())
	a.sseCancel = cancel
	go a.streamSSEToFrontend(sseCtx)
}

func (a *App) shutdown(ctx context.Context) {
	a.shutdownOnce.Do(func() {
		// Stop SSE stream
		if a.sseCancel != nil {
			a.sseCancel()
		}
		// Kill surge server if we spawned it
		if a.surgeProcess != nil && a.surgeProcess.Process != nil {
			_ = a.surgeProcess.Process.Signal(os.Interrupt)
			done := make(chan error, 1)
			go func() { done <- a.surgeProcess.Wait() }()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				_ = a.surgeProcess.Process.Kill()
			}
		}
	})
}

func (a *App) onSecondInstanceLaunch() {
	if a.ctx == nil {
		return
	}
	wailsRuntime.Show(a.ctx)
	wailsRuntime.WindowUnminimise(a.ctx)
	wailsRuntime.WindowShow(a.ctx)
}

// --------------------------------------------------------------------------
// Surge server management
// --------------------------------------------------------------------------

func (a *App) resolveSurgeBinary() string {
	// 1) Explicit override
	if envBin := strings.TrimSpace(os.Getenv("SURGE_BIN")); envBin != "" {
		if filepath.IsAbs(envBin) {
			if st, err := os.Stat(envBin); err == nil && !st.IsDir() {
				return envBin
			}
		} else if p, err := exec.LookPath(envBin); err == nil {
			return p
		}
	}

	homeDir, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(homeDir, "go", "bin", "surge"),
		"/usr/local/bin/surge",
		"/opt/homebrew/bin/surge",
		"/opt/local/bin/surge",
	}

	// 2) Known absolute locations (important for Finder/Dock launches)
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}

	// 3) PATH fallback
	if p, err := exec.LookPath("surge"); err == nil {
		return p
	}

	return ""
}

func (a *App) discoverToken() string {
	// Try to read token via surge CLI (from resolved binary)
	if a.surgeBin != "" {
		out, err := exec.Command(a.surgeBin, "token").Output()
		if err == nil {
			if t := strings.TrimSpace(string(out)); t != "" {
				return t
			}
		}
	}

	// Fallback: check common token file locations
	homeDir, _ := os.UserHomeDir()
	paths := []string{
		homeDir + "/.local/state/surge/token",
		homeDir + "/.local/share/surge/token",
		homeDir + "/.surge/token",
	}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			if t := strings.TrimSpace(string(data)); t != "" {
				return t
			}
		}
	}
	return ""
}

func (a *App) startSurgeServer() bool {
	if a.surgeBin == "" {
		wailsRuntime.LogError(a.ctx, "cannot start surge server: binary not resolved")
		return false
	}

	cmd := exec.Command(a.surgeBin, "server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		wailsRuntime.LogError(a.ctx, fmt.Sprintf("failed to start surge server (%s): %v", a.surgeBin, err))
		return false
	}

	a.surgeProcess = cmd
	wailsRuntime.LogInfo(a.ctx, fmt.Sprintf("Started surge server (PID %d) using %s", cmd.Process.Pid, a.surgeBin))
	return true
}

func (a *App) healthCheck() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, surgeBaseURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (a *App) waitForServer(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.healthCheck() {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

func (a *App) initAutostart() {
	exePath, err := os.Executable()
	if err != nil {
		wailsRuntime.LogError(a.ctx, fmt.Sprintf("failed to resolve executable path for autostart: %v", err))
		return
	}

	autostartID := "com.surge.desktop"
	a.autostartApp = &autostart.App{
		Name:        autostartID,
		DisplayName: "Surge",
		Exec:        []string{exePath},
	}
}

// GetStartupEnabled returns whether Launch At Login is enabled.
func (a *App) ensureAutostartApp() (*autostart.App, error) {
	if a.autostartApp == nil {
		a.initAutostart()
	}
	if a.autostartApp == nil {
		return nil, fmt.Errorf("autostart is not initialized")
	}
	return a.autostartApp, nil
}

func (a *App) GetStartupEnabled() bool {
	autostartApp, err := a.ensureAutostartApp()
	if err != nil {
		return false
	}
	return autostartApp.IsEnabled()
}

// SetStartupEnabled enables/disables Launch At Login.
func (a *App) SetStartupEnabled(enabled bool) error {
	autostartApp, err := a.ensureAutostartApp()
	if err != nil {
		return err
	}

	if enabled {
		return autostartApp.Enable()
	}
	if !autostartApp.IsEnabled() {
		return nil
	}
	return autostartApp.Disable()
}

func (a *App) iconModePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return ""
	}
	return filepath.Join(homeDir, ".config", "surge", "gui-icon-mode")
}

func normalizeIconMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case iconModeMenuBar:
		return iconModeMenuBar
	case iconModeDock:
		return iconModeDock
	default:
		return iconModeDock
	}
}

func (a *App) loadIconMode() {
	mode := iconModeDock
	p := a.iconModePath()
	if p != "" {
		if data, err := os.ReadFile(p); err == nil {
			mode = normalizeIconMode(string(data))
		}
	}

	if err := applyIconMode(mode); err != nil {
		wailsRuntime.LogWarning(a.ctx, fmt.Sprintf("apply icon mode failed: %v", err))
		a.iconMode = iconModeDock
		_ = applyIconMode(iconModeDock)
		return
	}
	a.iconMode = mode
}

func (a *App) saveIconMode(mode string) error {
	p := a.iconModePath()
	if p == "" {
		return fmt.Errorf("cannot resolve settings path")
	}

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(mode), 0o644)
}

func (a *App) GetIconMode() string {
	if a.iconMode == "" {
		a.loadIconMode()
	}
	return a.iconMode
}

func (a *App) SetIconMode(mode string) error {
	next := normalizeIconMode(mode)
	if err := applyIconMode(next); err != nil {
		return err
	}
	if err := a.saveIconMode(next); err != nil {
		return err
	}
	a.iconMode = next
	return nil
}

// --------------------------------------------------------------------------
// HTTP client helpers
// --------------------------------------------------------------------------

func (a *App) apiRequest(method, path string, body interface{}) (*http.Response, error) {
	url := surgeBaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}

	return http.DefaultClient.Do(req)
}

func (a *App) apiReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		if len(data) == 0 {
			return nil, fmt.Errorf("surge api %s returned %s", resp.Request.URL.Path, resp.Status)
		}
		return nil, fmt.Errorf("surge api %s returned %s: %s", resp.Request.URL.Path, resp.Status, strings.TrimSpace(string(data)))
	}
	return data, nil
}

func (a *App) apiGet(path string) ([]byte, error) {
	resp, err := a.apiRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return a.apiReadBody(resp)
}

func (a *App) apiPost(path string, body interface{}) ([]byte, error) {
	resp, err := a.apiRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	return a.apiReadBody(resp)
}

// --------------------------------------------------------------------------
// SSE Event streaming -> Wails EventsEmit -> Svelte
// --------------------------------------------------------------------------

func (a *App) streamSSEToFrontend(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		a.connectSSE(ctx)

		// Reconnect after 2 seconds on disconnect
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (a *App) connectSSE(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, surgeBaseURL+"/events", nil)
	if err != nil {
		return
	}
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	// Increase scanner buffer for large events
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var eventType string
	var data []byte

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			data = []byte(strings.TrimPrefix(line, "data: "))
		} else if line == "" && eventType != "" && data != nil {
			// Dispatch event
			a.dispatchSSEEvent(eventType, data)
			eventType = ""
			data = nil
		}
	}
}

func (a *App) dispatchSSEEvent(eventType string, data []byte) {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	wailsEventName := "surge:" + eventType
	wailsRuntime.EventsEmit(a.ctx, wailsEventName, payload)
}

// --------------------------------------------------------------------------
// Wails-bound methods (exposed to Svelte frontend)
// --------------------------------------------------------------------------

// DownloadItem is the JSON-friendly struct sent to the frontend
type DownloadItem struct {
	ID          string  `json:"id"`
	URL         string  `json:"url"`
	Filename    string  `json:"filename"`
	DestPath    string  `json:"dest_path"`
	TotalSize   int64   `json:"total_size"`
	Downloaded  int64   `json:"downloaded"`
	Progress    float64 `json:"progress"`
	Speed       float64 `json:"speed"`
	Status      string  `json:"status"`
	ETA         int64   `json:"eta"`
	Connections int     `json:"connections"`
	Error       string  `json:"error,omitempty"`
	TimeTaken   int64   `json:"time_taken"`
	AvgSpeed    float64 `json:"avg_speed"`
}

// ListDownloads returns all active downloads
func (a *App) ListDownloads() ([]DownloadItem, error) {
	data, err := a.apiGet("/list")
	if err != nil {
		return nil, err
	}
	var items []DownloadItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// AddDownload queues a new download
func (a *App) AddDownload(url, path, filename string) (string, error) {
	body := map[string]interface{}{
		"url":      url,
		"path":     path,
		"filename": filename,
	}
	data, err := a.apiPost("/download", body)
	if err != nil {
		return "", err
	}
	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	return result["id"], nil
}

// AddURL is a simpler binding: just provide a URL
func (a *App) AddURL(url string) (string, error) {
	return a.AddDownload(url, "", "")
}

func withDownloadID(path, id string) string {
	return path + "?" + url.Values{"id": []string{id}}.Encode()
}

// PauseDownload pauses a download
func (a *App) PauseDownload(id string) error {
	_, err := a.apiPost(withDownloadID("/pause", id), nil)
	return err
}

// ResumeDownload resumes a download
func (a *App) ResumeDownload(id string) error {
	_, err := a.apiPost(withDownloadID("/resume", id), nil)
	return err
}

// DeleteDownload removes a download
func (a *App) DeleteDownload(id string) error {
	resp, err := a.apiRequest(http.MethodDelete, withDownloadID("/delete", id), nil)
	if err != nil {
		return err
	}
	_, err = a.apiReadBody(resp)
	return err
}

// GetDownloadStatus returns the status of a single download
func (a *App) GetDownloadStatus(id string) (*DownloadItem, error) {
	data, err := a.apiGet(withDownloadID("/download", id))
	if err != nil {
		return nil, err
	}
	var item DownloadItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// GetHistory returns completed downloads
func (a *App) GetHistory() ([]json.RawMessage, error) {
	data, err := a.apiGet("/history")
	if err != nil {
		return nil, err
	}
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// OpenDirectoryDialog opens a native folder picker
func (a *App) OpenDirectoryDialog() (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Choose Download Directory",
	})
}

// IsServerRunning checks if the backend is alive
func (a *App) IsServerRunning() bool {
	return a.healthCheck()
}
