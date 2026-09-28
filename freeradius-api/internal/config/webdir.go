package config

import (
	"os"
	"path/filepath"
)

// resolveWebDir mencari direktori build frontend (SvelteKit adapter-static,
// ditandai oleh index.html). Urutan: WEB_DIR eksplisit -> ./web ->
// ../freeradius-web/build (CWD freeradius-api) -> freeradius-web/build
// (CWD repo root) -> <exe>/web.
// "" = tidak ditemukan, backend jalan mode API-only.
func resolveWebDir(explicit string) string {
	cands := []string{}
	if explicit != "" {
		cands = append(cands, explicit)
	}
	cands = append(cands, "web", "../freeradius-web/build", "freeradius-web/build")
	if exe, err := os.Executable(); err == nil {
		cands = append(cands, filepath.Join(filepath.Dir(exe), "web"))
	}
	for _, d := range cands {
		idx, err := filepath.Abs(filepath.Join(d, "index.html"))
		if err != nil {
			continue
		}
		if _, err := os.Stat(idx); err == nil {
			abs, _ := filepath.Abs(d)
			return abs
		}
	}
	return ""
}
