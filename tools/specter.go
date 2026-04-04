package tools

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

func RenderSpecter(w http.ResponseWriter, target string) {
	if target == "" { target = "127.0.0.1" }
	fmt.Fprintf(w, `
	<div class="card p-3 border-start border-danger border-4">
		<h6 class="fw-bold text-danger mb-3">🔍 SPECTER SCANNER</h6>
		<form method="POST" action="?admin=rdns&tool=specter&exec=true">
			<div class="input-group mb-3">
				<input type="text" name="target" class="form-control form-control-sm" value="%s">
				<button type="submit" class="btn btn-sm btn-danger">SCAN</button>
			</div>
		</form>
		<div class="console">`, target)

	ports := []string{"21", "22", "80", "443", "3306", "8080"}
	for _, port := range ports {
		address := net.JoinHostPort(target, port)
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err != nil {
			fmt.Fprintf(w, "<div>[PORT %s] <span style='color:red'>CLOSED</span></div>", port)
		} else {
			conn.Close()
			fmt.Fprintf(w, "<div>[PORT %s] <span style='color:#0f0'>OPEN</span></div>", port)
		}
	}
	fmt.Fprintf(w, `</div></div>`)
}
