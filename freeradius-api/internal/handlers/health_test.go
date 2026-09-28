package handlers

import "testing"

func TestHostMetricsReadable(t *testing.T) {
	// /proc + statfs parsing must yield sane, non-zero values on Linux.
	if _, total := memInfo(); total == 0 {
		t.Error("memInfo total = 0; want > 0")
	}
	if _, total := diskInfo("/"); total == 0 {
		t.Error("diskInfo total = 0; want > 0")
	}
	if got := freeradiusState(); got != "running" && got != "stopped" && got != "unknown" {
		t.Errorf("freeradiusState = %q; want running|stopped|unknown", got)
	}
}
