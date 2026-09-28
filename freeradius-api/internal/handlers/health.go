package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/appdb"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/database"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/radiox"
)

// HealthHandler reports server-level health for the System Health page:
// host metrics from /proc + statfs, the FreeRADIUS service state, the app DB
// counters and the RADIUS auth counters kept by the log tailer. No extra
// dependency: everything comes from the kernel and the existing tailer.
type HealthHandler struct {
	tail    *radiox.Tailer
	logPath string
	dbPath  string
	radAddr string
}

func NewHealthHandler(tail *radiox.Tailer, logPath, dbPath, radAddr string) *HealthHandler {
	return &HealthHandler{tail: tail, logPath: logPath, dbPath: dbPath, radAddr: radAddr}
}

func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	ok, fail := h.tail.Counts()
	ramUsed, ramTotal := memInfo()
	diskUsed, diskTotal := diskInfo("/")
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "System health retrieved successfully",
		"data": map[string]interface{}{
			"api": map[string]interface{}{
				"uptime_seconds": time.Since(startTime).Seconds(),
				"go_version":     goVersion(),
				"database":       databaseStatus(),
			},
			"host": map[string]interface{}{
				"hostname":         hostname(),
				"uptime_seconds":   hostUptime(),
				"cpu_count":        cpuCount(),
				"load1":            loadAvg()[0],
				"load5":            loadAvg()[1],
				"load15":           loadAvg()[2],
				"ram_used_bytes":   ramUsed,
				"ram_total_bytes":  ramTotal,
				"disk_used_bytes":  diskUsed,
				"disk_total_bytes": diskTotal,
			},
			"radius": map[string]interface{}{
				"status":    freeradiusState(),
				"log_path":  h.logPath,
				"test_addr": h.radAddr,
				"auth_ok":   ok,
				"auth_fail": fail,
			},
			"appdb": map[string]interface{}{"path": h.dbPath, "counts": appdb.Stats()},
		},
	})
}

// databaseStatus reports the primary RADIUS MySQL connectivity.
func databaseStatus() string {
	if database.DB != nil {
		if err := database.DB.Ping(); err == nil {
			return "connected"
		}
	}
	return "disconnected"
}

func goVersion() string { return runtime.Version() }

// ---- /proc readers -------------------------------------------------------

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// hostUptime reads seconds since boot from /proc/uptime.
func hostUptime() float64 {
	f, err := os.Open("/proc/uptime")
	if err != nil {
		return 0
	}
	defer f.Close()
	var up float64
	if _, err := fmt.Fscan(f, &up); err != nil {
		return 0
	}
	return up
}

func cpuCount() int { return runtime.NumCPU() }

func loadAvg() [3]float64 {
	var out [3]float64
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return out
	}
	fields := strings.Fields(string(data))
	for i := 0; i < 3 && i < len(fields); i++ {
		out[i], _ = strconv.ParseFloat(fields[i], 64)
	}
	return out
}

// memInfo returns used and total RAM in bytes (MemTotal - MemAvailable).
func memInfo() (used, total uint64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	var avail uint64
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			total = v * 1024
		case "MemAvailable:":
			avail = v * 1024
		}
	}
	if total > avail {
		used = total - avail
	}
	return used, total
}

// diskInfo returns used and total bytes for path.
func diskInfo(path string) (used, total uint64) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0
	}
	total = st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	used = total - free
	return used, total
}

// freeradiusState reports whether a freeradius/radiusd process is present.
// ponytail: scans /proc/comm instead of linking sd-bus or forking systemctl;
// no cgo dependency beyond the existing SQLite one.
func freeradiusState() string {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return "unknown"
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue
		}
		data, err := os.ReadFile("/proc/" + e.Name() + "/comm")
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(data))
		if name == "freeradius" || name == "radiusd" {
			return "running"
		}
	}
	return "stopped"
}
