package validation

import "testing"

func TestValidateUsernameParam(t *testing.T) {
	// Existing short usernames (path params) must be accepted.
	for _, ok := range []string{"admin", "user1", "testuser"} {
		if errs := ValidateUsernameParam(ok); len(errs) != 0 {
			t.Errorf("ValidateUsernameParam(%q) = %v, want no errors", ok, errs)
		}
	}
	// Empty and over-long are rejected.
	if errs := ValidateUsernameParam(""); len(errs) == 0 {
		t.Error("ValidateUsernameParam(\"\") want error")
	}
	if errs := ValidateUsernameParam(string(make([]byte, 65))); len(errs) == 0 {
		t.Error("ValidateUsernameParam(65 chars) want error")
	}
}

func TestValidateNASName(t *testing.T) {
	for _, ok := range []string{"core-router", "office_gw", "switch01"} {
		if errs := ValidateNASCreate(&NASInput{Name: ok, IP: "10.0.0.1", Secret: "secret123"}); len(errs) != 0 {
			t.Errorf("ValidateNASCreate(name=%q) = %v, want no errors", ok, errs)
		}
	}
	// A space or slash is still rejected.
	for _, bad := range []string{"core router", "a/b"} {
		if errs := ValidateNASCreate(&NASInput{Name: bad, IP: "10.0.0.1", Secret: "secret123"}); len(errs) == 0 {
			t.Errorf("ValidateNASCreate(name=%q) want error", bad)
		}
	}
}
