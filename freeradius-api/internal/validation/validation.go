package validation

import (
	"fmt"
	"net"
	"strings"
)

// FieldError mirrors the Node/Joi error shape: {field, message}.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var validNASTypes = map[string]bool{
	"cisco": true, "computone": true, "livingston": true, "juniper": true,
	"max40xx": true, "multitech": true, "netserver": true, "pathras": true,
	"patton": true, "portslave": true, "tc": true, "usrhiper": true, "other": true,
}

func isAlnum(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

type NASInput struct {
	Name        string
	IP          string
	Secret      string
	Type        string
	Ports       *int
	Community   string
	Description string
	HasType     bool
	HasPorts    bool
	HasComm     bool
	HasDesc     bool
}

// ValidateNASCreate enforces the same rules as the Node Joi schema.
func ValidateNASCreate(in *NASInput) []FieldError {
	var errs []FieldError
	if !isAlnum(in.Name) {
		errs = append(errs, FieldError{"name", "Name must contain only alphanumeric characters"})
	} else if len(in.Name) < 3 {
		errs = append(errs, FieldError{"name", "Name must be at least 3 characters long"})
	} else if len(in.Name) > 30 {
		errs = append(errs, FieldError{"name", "Name must not exceed 30 characters"})
	}
	if in.IP == "" || net.ParseIP(in.IP) == nil {
		errs = append(errs, FieldError{"ip", "Must be a valid IP address"})
	}
	if len(in.Secret) < 8 {
		errs = append(errs, FieldError{"secret", "Secret must be at least 8 characters long"})
	} else if len(in.Secret) > 100 {
		errs = append(errs, FieldError{"secret", "Secret must not exceed 100 characters"})
	}
	t := in.Type
	if t == "" {
		t = "other"
	}
	if !validNASTypes[t] {
		errs = append(errs, FieldError{"type", fmt.Sprintf("Type must be one of %s", strings.Join(nasTypeList(), ", "))})
	}
	if in.Ports != nil && (*in.Ports < 1 || *in.Ports > 65535) {
		errs = append(errs, FieldError{"ports", "Ports must be between 1 and 65535"})
	}
	if len(in.Community) > 50 {
		errs = append(errs, FieldError{"community", "Community must not exceed 50 characters"})
	}
	if len(in.Description) > 200 {
		errs = append(errs, FieldError{"description", "Description must not exceed 200 characters"})
	}
	return errs
}

func ValidateNASUpdate(in *NASInput) []FieldError {
	var errs []FieldError
	if in.Name != "" {
		if !isAlnum(in.Name) {
			errs = append(errs, FieldError{"name", "Name must contain only alphanumeric characters"})
		} else if len(in.Name) < 3 {
			errs = append(errs, FieldError{"name", "Name must be at least 3 characters long"})
		} else if len(in.Name) > 30 {
			errs = append(errs, FieldError{"name", "Name must not exceed 30 characters"})
		}
	}
	if in.IP != "" && net.ParseIP(in.IP) == nil {
		errs = append(errs, FieldError{"ip", "Must be a valid IP address"})
	}
	if in.Secret != "" {
		if len(in.Secret) < 8 {
			errs = append(errs, FieldError{"secret", "Secret must be at least 8 characters long"})
		} else if len(in.Secret) > 100 {
			errs = append(errs, FieldError{"secret", "Secret must not exceed 100 characters"})
		}
	}
	if in.HasType && !validNASTypes[in.Type] {
		errs = append(errs, FieldError{"type", "Invalid NAS type"})
	}
	if in.HasPorts && in.Ports != nil && (*in.Ports < 1 || *in.Ports > 65535) {
		errs = append(errs, FieldError{"ports", "Ports must be between 1 and 65535"})
	}
	if in.HasComm && len(in.Community) > 50 {
		errs = append(errs, FieldError{"community", "Community must not exceed 50 characters"})
	}
	if in.HasDesc && len(in.Description) > 200 {
		errs = append(errs, FieldError{"description", "Description must not exceed 200 characters"})
	}
	return errs
}

func nasTypeList() []string {
	return []string{"cisco", "computone", "livingston", "juniper", "max40xx", "multitech", "netserver", "pathras", "patton", "portslave", "tc", "usrhiper", "other"}
}

func ValidateUserCreate(username, password string) []FieldError {
	var errs []FieldError
	if len(username) < 6 {
		errs = append(errs, FieldError{"user", "Username must be at least 6 characters long"})
	}
	if len(password) < 6 {
		errs = append(errs, FieldError{"password", "Password must be at least 6 characters long"})
	}
	return errs
}

func ValidateUserUpdate(password string, hasPassword bool) []FieldError {
	var errs []FieldError
	if hasPassword && len(password) < 6 {
		errs = append(errs, FieldError{"password", "Password must be at least 6 characters long"})
	}
	return errs
}

func ValidateUsernameParam(username string) []FieldError {
	var errs []FieldError
	if len(username) < 6 {
		errs = append(errs, FieldError{"username", "Username must be at least 6 characters long"})
	} else if len(username) > 64 {
		errs = append(errs, FieldError{"username", "Username must not exceed 64 characters"})
	}
	return errs
}
