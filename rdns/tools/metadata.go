package tools

import (
	"fmt"
	"net/http"
	"strings"
)

func RenderMetaStripper(w http.ResponseWriter, lang string, isProcessed bool, fileData map[string]string, rows []string) {
	fmt.Fprintf(w, `<div class="card border-start-success">
<h6 class="fw-bold mb-3" style="color:var(--success)">METADATA ENGINE</h6>
<form method="POST" enctype="multipart/form-data" action="?admin=rdns&tool=metadata">
<div class="upload-area">
<label class="mb-2">Upload Gambar (JPEG / PNG / GIF) - Max 20MB</label>
<input type="file" name="imagefile" class="form-control mb-2" accept="image/*" required>
<button type="submit" class="btn btn-success w-100 fw-bold">ANALYSIS &amp; STRIP EXIF</button>
</div>
</form>`)

	if isProcessed && fileData != nil {
		ip := fileData["ip"]
		if ip == "" { ip = "unknown" }
		fmt.Fprintf(w, `<div class="report-box">
<div class="report-title">DUMP REPORT:</div>
<div class="report-ip">&gt; SOURCE IP  : %s</div>
<div style="color:#ccc">&gt; FILE NAME  : %s</div>
<div style="color:#ccc">&gt; FILE SIZE  : %s</div>
<div style="color:#ccc">&gt; TIMESTAMP  : %s</div>
<br>
<div class="report-warn">[!] EXTRACTING EXIF METADATA...</div>
<div class="report-err">[X] EXIF DATA STRIPPED!</div>
<div class="report-ok">[OK] IMAGE CLEANED SUCCESSFULLY.</div>
</div>
<a href="?admin=rdns&tool=metadata&action=download" class="btn btn-primary w-100 fw-bold mb-3">DOWNLOAD SECURE IMAGE</a>`,
			ip, fileData["name"], fileData["size"], fileData["time"])
	}

	fmt.Fprintf(w, `<h6 class="fw-bold mt-2 mb-2" style="color:var(--text2);font-size:11px">RECENT ACTIVITY</h6>`)

	var valid []string
	for _, r := range rows {
		if strings.TrimSpace(r) != "" { valid = append(valid, r) }
	}

	if len(valid) == 0 {
		fmt.Fprintf(w, `<div class="empty-state">Belum ada gambar yang diproses.</div>`)
	} else {
		fmt.Fprintf(w, `<div class="table-responsive rounded border" style="max-height:200px;overflow-y:auto">
<table class="table table-sm">
<thead><tr><th>WAKTU</th><th>IP</th><th>FILENAME</th><th>SIZE</th></tr></thead>
<tbody>`)
		for i := len(valid) - 1; i >= 0; i-- {
			p := strings.SplitN(valid[i], "|", 5)
			ts, ip, fname, size := "-", "-", "-", "-"
			if len(p) > 0 { ts = strings.TrimSpace(p[0]) }
			if len(p) > 1 {
				ip = strings.TrimSpace(p[1])
				if ip == "::1" { ip = "127.0.0.1" }
			}
			if len(p) > 2 { fname = strings.TrimSpace(p[2]) }
			if len(p) > 3 { size = strings.TrimSpace(p[3]) }
			fmt.Fprintf(w,
				`<tr><td style="font-size:10px;white-space:nowrap;color:var(--text3)">%s</td>`+
				`<td><code style="color:var(--primary);font-size:10px">%s</code></td>`+
				`<td style="color:var(--warning);font-size:10px;word-break:break-all">%s</td>`+
				`<td style="color:var(--text3);font-size:10px">%s</td></tr>`,
				ts, ip, fname, size)
		}
		fmt.Fprintf(w, `</tbody></table></div>`)
	}
	fmt.Fprintf(w, `</div>`)
}
