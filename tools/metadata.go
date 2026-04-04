package tools

import (
	"fmt"
	"net/http"
)

func RenderMetaStripper(w http.ResponseWriter, lang string, isProcessed bool, fileData map[string]string, rows []string) {
	fmt.Fprintf(w, `
	<div class="card p-3 shadow-sm border-start border-success border-4">
		<h6 class="fw-bold text-success mb-3">🖼️ METADATA ENGINE</h6>
		<form method="POST" enctype="multipart/form-data" action="?admin=rdns&tool=metadata">
			<div class="p-4 text-center rounded border border-2 border-dashed bg-light mb-3">
				<input type="file" name="imagefile" class="form-control form-control-sm mb-3" accept="image/*" required>
				<button type="submit" class="btn btn-success btn-sm w-100 fw-bold">ANALYSIS & STRIP</button>
			</div>
		</form>`)

	// JIKA GAMBAR SUDAH DI-UPLOAD, TAMPILKAN DATA DINAMIS!
	if isProcessed && fileData != nil {
		fmt.Fprintf(w, `
		<div class="p-3 border rounded bg-dark text-success mb-3" style="font-size:11px; font-family:monospace;">
			<div class="text-white border-bottom mb-2 pb-1">DUMP REPORT (LIVE):</div>
			<div class="text-primary">> TARGET IP: %s</div>
			<div>> FILE NAME: %s</div>
			<div>> FILE SIZE: %s</div>
			<div>> TIMESTAMP: %s</div>
			<br>
			<div class="text-warning">[!] EXTRACTING EXIF...</div>
			<div class="text-danger">[X] EXIF DATA DESTROYED!</div>
			<div class="text-success mt-2">[OK] PAYLOAD CLEANED SUCCESSFULLY.</div>
		</div>
		<a href="?admin=rdns&tool=metadata&action=download" class="btn btn-primary btn-sm w-100 fw-bold mb-3">DOWNLOAD SECURE IMAGE</a>`,
			fileData["ip"], fileData["name"], fileData["size"], fileData["time"])
	}

	// TABEL LOG HISTORY
	fmt.Fprintf(w, `<h6 class="fw-bold mt-4 mb-2 small">⏱️ RECENT ACTIVITY</h6>`)
	if len(rows) == 0 || rows[0] == "" {
		fmt.Fprintf(w, `<div class="p-3 text-center text-muted small border rounded">No data stripped yet.</div>`)
	} else {
		fmt.Fprintf(w, `<div class="table-responsive bg-white rounded border" style="max-height:200px">
			<table class="table table-sm small mb-0"><tbody>`)
		for i := len(rows) - 1; i >= 0; i-- {
			if rows[i] == "" { continue }
			fmt.Fprintf(w, `<tr><td><code>%s</code></td></tr>`, rows[i])
		}
		fmt.Fprintf(w, `</tbody></table></div>`)
	}
	fmt.Fprintf(w, `</div>`)
}
