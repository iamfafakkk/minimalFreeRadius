package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/config"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/models"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
	"layeh.com/radius/rfc3576"
)

// DefaultNASCoAPort is the RFC 5176 CoA/Disconnect port. MikroTik, Cisco and
// most NAS use 3799 unless overridden per device (nas.ports).
const DefaultNASCoAPort = 3799

type NASHandler struct{ cfg *config.Config }

func NewNASHandler(cfg *config.Config) *NASHandler { return &NASHandler{cfg: cfg} }

// coaTarget strips a CIDR mask / zone from nasname (which is stored as an IP or
// network) and returns the UDP host:port used for CoA/Disconnect.
func coaTarget(n *models.NAS) (string, int) {
	host := n.IP
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	if i := strings.IndexByte(host, '%'); i >= 0 {
		host = host[:i]
	}
	port := DefaultNASCoAPort
	if n.Ports != nil && *n.Ports > 0 {
		port = *n.Ports
	}
	return host, port
}

// signMessageAuthenticator fills the Message-Authenticator attribute (HMAC-MD5
// over the packet, RFC 3579/5176). layeh.com/radius does not compute it, and
// FreeRADIUS 3.0.26 drops Access/CoA requests that omit it. Must be called
// before Exchange (Encode then finalizes the CoA Request Authenticator).
func signMessageAuthenticator(p *radius.Packet, secret string) error {
	if err := rfc2869.MessageAuthenticator_Set(p, make([]byte, 16)); err != nil {
		return err
	}
	wire, err := p.MarshalBinary()
	if err != nil {
		return err
	}
	// RFC 5176: CoA/Disconnect compute the MAC with a zeroed Request Authenticator.
	if p.Code == radius.CodeCoARequest || p.Code == radius.CodeDisconnectRequest {
		for i := 4; i < 20; i++ {
			wire[i] = 0
		}
	}
	mac := hmac.New(md5.New, []byte(secret))
	mac.Write(wire)
	return rfc2869.MessageAuthenticator_Set(p, mac.Sum(nil))
}

type nasTestResult struct {
	Action    string `json:"action"`
	Target    string `json:"target"`
	Status    string `json:"status"` // ack | nak | accept | reject | challenge | timeout | error | unexpected
	Message   string `json:"message"`
	Reply     string `json:"reply,omitempty"`
	LatencyMs int64  `json:"latency_ms"`
	SecretOK  *bool  `json:"secret_ok,omitempty"`
}

// Test sends a Disconnect-Request (default) or CoA-Request to the NAS itself,
// using the NAS secret and CoA port. A Disconnect-ACK means the device is
// reachable AND the shared secret is correct.
func (h *NASHandler) Test(w http.ResponseWriter, r *http.Request) {
	id, ok := PathID(r)
	if !ok {
		WriteErr(w, http.StatusBadRequest, "ID must be greater than 0")
		return
	}
	n, err := models.GetNASByID(id)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if n == nil {
		WriteErr(w, http.StatusNotFound, "NAS not found")
		return
	}

	var body struct {
		Action   string `json:"action"` // "disconnect" (default) or "coa"
		Username string `json:"username"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	action := strings.ToLower(strings.TrimSpace(body.Action))
	code := radius.CodeDisconnectRequest
	if action == "coa" {
		code = radius.CodeCoARequest
	} else {
		action = "disconnect"
	}

	host, port := coaTarget(n)
	target := net.JoinHostPort(host, strconv.Itoa(port))
	if n.Secret == "" {
		WriteErr(w, http.StatusBadRequest, "NAS has no shared secret")
		return
	}

	packet := radius.New(code, []byte(n.Secret))
	if body.Username != "" {
		_ = rfc2865.UserName_SetString(packet, body.Username)
	}
	if err := signMessageAuthenticator(packet, n.Secret); err != nil {
		WriteErr(w, http.StatusBadRequest, "Failed to sign request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	start := time.Now()
	resp, err := radius.Exchange(ctx, packet, target)
	latency := time.Since(start).Milliseconds()

	res := nasTestResult{Action: action, Target: target, LatencyMs: latency}
	if err != nil {
		res.Status = "timeout"
		res.Message = "No reply from NAS (unreachable, wrong CoA port, or CoA disabled). " + err.Error()
		WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": res.Message, "data": res})
		return
	}

	secretOK := true
	res.SecretOK = &secretOK
	switch resp.Code {
	case radius.CodeDisconnectACK, radius.CodeCoAACK:
		res.Status = "ack"
		res.Message = "Success: NAS accepted the request. Device reachable and shared secret is correct."
	case radius.CodeDisconnectNAK, radius.CodeCoANAK:
		res.Status = "nak"
		res.Message = "NAS rejected the request (reachable, secret OK)."
		if cause := rfc3576.ErrorCause_Get(resp); cause != 0 {
			res.Reply = cause.String()
			res.Message += " Cause: " + cause.String()
		}
	default:
		res.Status = "unexpected"
		res.Message = "Unexpected reply: " + resp.Code.String()
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": res.Message, "data": res})
}

// TestAuth sends an Access-Request to the local FreeRADIUS server and reports
// Accept/Reject. It uses the client secret for the request source (default
// 127.0.0.1 -> "testing123"), NOT the NAS secret, because the request comes
// from the panel host rather than the NAS.
func (h *NASHandler) TestAuth(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		NASID    int64  `json:"nas_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteErr(w, http.StatusBadRequest, "Invalid JSON in request body")
		return
	}
	if strings.TrimSpace(body.Username) == "" || body.Password == "" {
		WriteErr(w, http.StatusBadRequest, "Username and password are required")
		return
	}
	if h.cfg.RadiusTestSecret == "" {
		WriteErr(w, http.StatusBadRequest, "Local client secret is not configured (RADIUS_TEST_SECRET)")
		return
	}

	packet := radius.New(radius.CodeAccessRequest, []byte(h.cfg.RadiusTestSecret))
	_ = rfc2865.UserName_SetString(packet, body.Username)
	if err := rfc2865.UserPassword_SetString(packet, body.Password); err != nil {
		WriteErr(w, http.StatusBadRequest, "Failed to encode password")
		return
	}
	// Present the NAS identity to the server (optional, informational).
	if body.NASID > 0 {
		if n, _ := models.GetNASByID(body.NASID); n != nil {
			_ = rfc2865.NASIdentifier_SetString(packet, n.Name)
		}
	}
	if err := signMessageAuthenticator(packet, h.cfg.RadiusTestSecret); err != nil {
		WriteErr(w, http.StatusBadRequest, "Failed to sign request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	start := time.Now()
	resp, err := radius.Exchange(ctx, packet, h.cfg.RadiusTestAddr)
	latency := time.Since(start).Milliseconds()

	res := nasTestResult{Action: "auth", Target: h.cfg.RadiusTestAddr, LatencyMs: latency}
	if err != nil {
		res.Status = "error"
		res.Message = "No reply from RADIUS server. " + err.Error()
		WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": res.Message, "data": res})
		return
	}
	res.Reply = rfc2865.ReplyMessage_GetString(resp)
	switch resp.Code {
	case radius.CodeAccessAccept:
		res.Status = "accept"
		res.Message = "Access-Accept: credentials are valid."
	case radius.CodeAccessReject:
		res.Status = "reject"
		res.Message = "Access-Reject: invalid credentials."
	case radius.CodeAccessChallenge:
		res.Status = "challenge"
		res.Message = "Access-Challenge returned."
	default:
		res.Status = "unexpected"
		res.Message = "Unexpected reply: " + resp.Code.String()
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": res.Message, "data": res})
}
