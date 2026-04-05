package tools

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type portDef struct{ num, name string }

var CommonPorts = []portDef{
	{"21","FTP"},{"22","SSH"},{"23","TELNET"},{"25","SMTP"},
	{"53","DNS"},{"80","HTTP"},{"110","POP3"},{"143","IMAP"},
	{"443","HTTPS"},{"445","SMB"},{"3306","MYSQL"},{"3389","RDP"},
	{"5432","POSTGRESQL"},{"6379","REDIS"},{"8080","HTTP-ALT"},
	{"8443","HTTPS-ALT"},{"27017","MONGODB"},
}

func RenderSpecter(w http.ResponseWriter, target string) {
	fmt.Fprintf(w, `<div class="card border-start-danger"><h6 class="fw-bold mb-3" style="color:var(--danger)">SPECTER SCANNER</h6><form method="POST" action="?admin=rdns&tool=specter&exec=true"><label>TARGET IP / DOMAIN</label><div class="input-group mb-3"><input type="text" name="target" class="form-control" value="%s" placeholder="127.0.0.1 atau example.com"><button type="submit" class="btn btn-danger fw-bold">SCAN</button></div></form>`, target)
	if target != "" {
		openCount := 0
		fmt.Fprintf(w, `<div class="console" style="height:320px;overflow-y:auto"><div style="color:#fff;margin-bottom:8px">[*] SPECTER SCAN - TARGET: <span style="color:#ff5252">%s</span></div><div style="color:#666;margin-bottom:12px">[*] Scanning %d ports...</div>`, target, len(CommonPorts))
		for _, p := range CommonPorts {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, p.num), 700*time.Millisecond)
			if err != nil {
				fmt.Fprintf(w, `<div>[<span style="color:#444">%s/%s</span>] <span style="color:#555">CLOSED</span></div>`, p.num, p.name)
			} else {
				conn.Close()
				openCount++
				fmt.Fprintf(w, `<div>[<span style="color:#00e676;font-weight:bold">%s/%s</span>] <span style="color:#00e676;font-weight:bold">OPEN</span></div>`, p.num, p.name)
			}
		}
		fmt.Fprintf(w, `<div style="margin-top:12px;color:#ffd740">[DONE] Scan selesai. <span style="color:#00e676">%d port terbuka</span> dari %d.</div></div>`, openCount, len(CommonPorts))
	} else {
		fmt.Fprintf(w, `<div class="empty-state">Masukkan target IP atau domain untuk memulai scan.</div>`)
	}
	fmt.Fprintf(w, `</div>`)
}
