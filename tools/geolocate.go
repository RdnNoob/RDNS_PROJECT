package tools

import (
	"fmt"
	"net/http"
	"strings"
)

func RenderGeoTable(w http.ResponseWriter, rows []string, full bool) {
	if len(rows) == 0 || rows[0] == "" {
		fmt.Fprintf(w, `<div class="p-4 text-center text-muted small">Belum ada data target.</div>`)
		return
	}
	fmt.Fprintf(w, `<div class="table-responsive"><table class="table table-sm small"><thead><tr><th>JAM</th><th>KOORDINAT</th><th class="text-end">MAP</th></tr></thead><tbody>`)
	
	start := 0
	if !full && len(rows) > 10 { start = len(rows) - 10 }
	
	for i := len(rows) - 1; i >= start; i-- {
		c := strings.Split(rows[i], "|")
		if len(c) < 3 { continue }
		fmt.Fprintf(w, `<tr><td>%s</td><td class="text-danger"><code>%s</code></td><td class="text-end"><a href="https://www.google.com/maps?q=%s" target="_blank" class="btn btn-xs btn-primary py-0 px-2" style="font-size:10px">GO</a></td></tr>`, c[0], c[2], c[2])
	}
	fmt.Fprintf(w, `</tbody></table></div>`)
}
