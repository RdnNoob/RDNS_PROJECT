package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const ngrokTokenFile = "database/ngrok.token"

var (
	tunnelMu    sync.Mutex
	tunnelCmd   *exec.Cmd
	tunnelLogMu sync.Mutex
	tunnelLogs  []string
)

// Domain resmi tiap provider tunnel
var tunnelDomains = map[string][]string{
	"cloudflare":   {".trycloudflare.com"},
	"ngrok":        {".ngrok.io", ".ngrok-free.app", ".ngrok.app"},
	"serveo":       {".serveo.net"},
	"localhostrun": {".lhr.life", ".localhost.run"},
}

func appendTunnelLog(line string) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return
	}
	tunnelLogMu.Lock()
	tunnelLogs = append(tunnelLogs, line)
	if len(tunnelLogs) > 200 {
		tunnelLogs = tunnelLogs[len(tunnelLogs)-200:]
	}
	tunnelLogMu.Unlock()
}

func getTunnelLogs() []string {
	tunnelLogMu.Lock()
	defer tunnelLogMu.Unlock()
	r := make([]string, len(tunnelLogs))
	copy(r, tunnelLogs)
	return r
}

func clearTunnelLogs() {
	tunnelLogMu.Lock()
	tunnelLogs = nil
	tunnelLogMu.Unlock()
}

// findURLForProvider mencari URL yang domainnya cocok dengan provider
func findURLForProvider(text, provider string) string {
	domains := tunnelDomains[provider]
	from := 0
	for {
		idx := strings.Index(text[from:], "https://")
		if idx < 0 {
			return ""
		}
		idx += from
		end := idx
		for end < len(text) {
			ch := text[end]
			if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' ||
				ch == '"' || ch == '\'' || ch == '|' || ch == '>' ||
				ch == '<' || ch == '}' || ch == ')' || ch == ']' || ch == '\\' {
				break
			}
			end++
		}
		url := strings.TrimRight(text[idx:end], ".,;:")
		if len(url) > 15 {
			matched := false
			for _, d := range domains {
				if strings.Contains(url, d) {
					matched = true
					break
				}
			}
			if matched {
				return url
			}
		}
		from = end
		if from >= len(text) {
			return ""
		}
	}
}

// scanReader membaca output tunnel, log tiap baris, cari URL sesuai provider
func scanReader(rd interface{ Read([]byte) (int, error) }, provider string) string {
	var buf strings.Builder
	var lineBuf strings.Builder
	chunk := make([]byte, 4096)
	for {
		n, err := rd.Read(chunk)
		if n > 0 {
			text := string(chunk[:n])
			buf.WriteString(text)
			for _, ch := range text {
				if ch == '\n' {
					appendTunnelLog(lineBuf.String())
					lineBuf.Reset()
				} else {
					lineBuf.WriteRune(ch)
				}
			}
			if url := findURLForProvider(buf.String(), provider); url != "" {
				appendTunnelLog("[TUNNEL URL] " + url)
				return url
			}
		}
		if err != nil {
			break
		}
	}
	if lineBuf.Len() > 0 {
		appendTunnelLog(lineBuf.String())
	}
	return ""
}

func setTunnelURL(url string) {
	if url == "" {
		return
	}
	sessLock.Lock()
	updated := false
	if session.TunnelURL == "Generating..." {
		session.TunnelURL = url
		updated = true
	}
	sessLock.Unlock()
	if updated {
		saveSession()
		fmt.Println("[TUNNEL] URL:", url)
	}
}

func saveNgrokToken(token string) {
	if token != "" {
		os.WriteFile(ngrokTokenFile, []byte(strings.TrimSpace(token)), 0600)
	}
}

func loadNgrokToken() string {
	data, err := os.ReadFile(ngrokTokenFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func startTunnel(provider string) {
	killTunnel()
	clearTunnelLogs()
	appendTunnelLog("[*] Memulai tunnel: " + provider)
	tunnelMu.Lock()
	var cmd *exec.Cmd
	switch provider {
	case "ngrok":
		token := loadNgrokToken()
		if token != "" {
			appendTunnelLog("[*] Mengkonfigurasi ngrok authtoken...")
			cfgCmd := exec.Command("ngrok", "config", "add-authtoken", token)
			out, err := cfgCmd.CombinedOutput()
			if err != nil {
				appendTunnelLog("[!] Authtoken error: " + err.Error())
			} else {
				appendTunnelLog("[+] Authtoken OK: " + strings.TrimSpace(string(out)))
			}
		} else {
			appendTunnelLog("[!] Ngrok authtoken tidak ditemukan. Masukkan token di form config.")
		}
		cmd = exec.Command("ngrok", "http", "8080")
	case "cloudflare":
		cmd = exec.Command("cloudflared", "tunnel", "--url",
			fmt.Sprintf("http://localhost%s", port))
	case "serveo":
		cmd = exec.Command("ssh",
			"-o", "StrictHostKeyChecking=no",
			"-o", "ServerAliveInterval=30",
			"-o", "ConnectTimeout=20",
			"-R", "80:localhost:8080", "serveo.net")
	case "localhostrun":
		cmd = exec.Command("ssh",
			"-o", "StrictHostKeyChecking=no",
			"-o", "ServerAliveInterval=30",
			"-o", "ConnectTimeout=20",
			"-R", "80:localhost:8080", "localhost.run")
	default:
		cmd = exec.Command("cloudflared", "tunnel", "--url",
			fmt.Sprintf("http://localhost%s", port))
	}
	tunnelCmd = cmd
	tunnelMu.Unlock()

	// Timeout fallback 90 detik
	go func() {
		time.Sleep(90 * time.Second)
		sessLock.Lock()
		timedOut := session.TunnelURL == "Generating..."
		if timedOut {
			session.TunnelURL = "[Timeout: " + provider + " tidak menghasilkan URL]"
		}
		sessLock.Unlock()
		if timedOut {
			appendTunnelLog("[!] Timeout: tidak ada URL dari " + provider)
			saveSession()
		}
	}()

	if provider == "ngrok" {
		if err := cmd.Start(); err != nil {
			msg := "[Error ngrok: " + err.Error() + "]"
			appendTunnelLog(msg)
			sessLock.Lock()
			session.TunnelURL = msg
			sessLock.Unlock()
			saveSession()
			return
		}
		appendTunnelLog("[*] Ngrok berjalan, polling API port 4040...")
		go func() {
			for i := 0; i < 60; i++ {
				time.Sleep(1 * time.Second)
				if url := getNgrokURL(); url != "" {
					appendTunnelLog("[TUNNEL URL] " + url)
					setTunnelURL(url)
					return
				}
			}
		}()
		return
	}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		msg := "[Error: " + err.Error() + "]"
		appendTunnelLog(msg)
		sessLock.Lock()
		session.TunnelURL = msg
		sessLock.Unlock()
		saveSession()
		return
	}
	appendTunnelLog("[*] Proses tunnel berjalan, membaca output...")
	go func() { setTunnelURL(scanReader(stdout, provider)) }()
	go func() { setTunnelURL(scanReader(stderr, provider)) }()
}

func getNgrokURL() string {
	resp, err := http.Get("http://localhost:4040/api/tunnels")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var result struct {
		Tunnels []struct {
			PublicURL string `json:"public_url"`
		} `json:"tunnels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	for _, t := range result.Tunnels {
		if strings.HasPrefix(t.PublicURL, "https://") &&
			!strings.Contains(t.PublicURL, "localhost") {
			return t.PublicURL
		}
	}
	return ""
}

func killTunnel() {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()
	if tunnelCmd != nil && tunnelCmd.Process != nil {
		tunnelCmd.Process.Kill()
		tunnelCmd.Wait()
		tunnelCmd = nil
	}
}

func getTunnelURL() string {
	sessLock.RLock()
	defer sessLock.RUnlock()
	if session.TunnelURL == "" {
		return "[Tunnel tidak aktif]"
	}
	return session.TunnelURL
}
