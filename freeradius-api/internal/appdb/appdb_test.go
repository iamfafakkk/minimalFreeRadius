package appdb

import (
	"path/filepath"
	"testing"
)

// TestIdempotentSeed verifies that opening the DB and seeding the admin twice
// (i.e. restarting the server) never duplicates or overwrites the row.
func TestIdempotentSeed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")

	if err := Open(path); err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if err := SeedAdmin("admin", "admin123!"); err != nil {
		t.Fatalf("first SeedAdmin: %v", err)
	}
	id1, role1, ok1 := Authenticate("admin", "admin123!")
	if !ok1 {
		t.Fatal("authenticate after first seed failed")
	}
	Close()

	// Simulate a restart with a *different* configured password: the existing
	// row must be kept as-is (INSERT OR IGNORE).
	if err := Open(path); err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer Close()
	if err := SeedAdmin("admin", "different-password"); err != nil {
		t.Fatalf("second SeedAdmin: %v", err)
	}

	users, err := ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("want exactly 1 app user after re-seed, got %d", len(users))
	}
	id2, role2, ok2 := Authenticate("admin", "admin123!")
	if !ok2 || id1 != id2 || role1 != role2 {
		t.Fatalf("existing credentials changed on re-seed: ok=%v id=%d/%d role=%q/%q", ok2, id1, id2, role1, role2)
	}

	// Activity + login logging round-trips.
	if err := RecordLogin(id2, "admin", "127.0.0.1", "test", true); err != nil {
		t.Fatalf("RecordLogin: %v", err)
	}
	if err := RecordActivity(id2, "admin", "create", "POST", "/api/v1/nas/", 201, "127.0.0.1", "test"); err != nil {
		t.Fatalf("RecordActivity: %v", err)
	}
	logins, _ := RecentLogins(10)
	acts, _ := RecentActivity(10)
	if len(logins) != 1 || !logins[0].Success {
		t.Fatalf("want 1 successful login, got %+v", logins)
	}
	if len(acts) != 1 || acts[0].Action != "create" || acts[0].Status != 201 {
		t.Fatalf("want 1 create activity, got %+v", acts)
	}
	if s := Stats(); s["users"] != 1 || s["logins"] != 1 || s["activity"] != 1 {
		t.Fatalf("unexpected stats: %+v", s)
	}
}
