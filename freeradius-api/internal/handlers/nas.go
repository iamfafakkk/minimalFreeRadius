package handlers

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/models"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/validation"
)

func (h *NASHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := QueryPageLimit(r)
	search := r.URL.Query().Get("search")
	items, err := models.GetAllNAS(search)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	paged, total, pages := Paginate(items, page, limit)
	// Keep legacy shape: data + count, plus pagination for swagger compat.
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "NAS entries retrieved successfully",
		"data": paged, "count": len(paged),
		"pagination": map[string]interface{}{"page": page, "limit": limit, "total": total, "pages": pages},
	})
}

func (h *NASHandler) Stats(w http.ResponseWriter, r *http.Request) {
	c, err := models.CountNAS()
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "NAS statistics retrieved successfully",
		"data": map[string]interface{}{"total_nas": c},
	})
}

func (h *NASHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "NAS retrieved successfully", "data": n,
	})
}

func (h *NASHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		IP          string `json:"ip"`
		Secret      string `json:"secret"`
		Type        string `json:"type"`
		Ports       *int   `json:"ports"`
		Community   string `json:"community"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteErr(w, http.StatusBadRequest, "Invalid JSON in request body")
		return
	}
	in := validation.NASInput{Name: body.Name, IP: body.IP, Secret: body.Secret, Type: body.Type, Community: body.Community, Description: body.Description}
	if body.Ports != nil {
		in.Ports = body.Ports
	}
	if errs := validation.ValidateNASCreate(&in); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	typ := in.Type
	if typ == "" {
		typ = "other"
	}
	ports := 1812
	if in.Ports != nil {
		ports = *in.Ports
	}
	exists, err := models.NASExists(in.Name, in.IP, nil)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if exists {
		WriteErr(w, http.StatusConflict, "NAS with this name or IP already exists")
		return
	}
	n, err := models.CreateNAS(models.NASCreate{
		Name: in.Name, IP: in.IP, Secret: in.Secret, Type: typ,
		Ports: ports, Community: in.Community, Description: in.Description,
	})
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	reloadOut, reloadOK := h.reloadRadius()
	msg := "NAS created successfully. FreeRADIUS restarted to load the new client."
	if !reloadOK {
		msg = "NAS created, but FreeRADIUS could not be restarted (" + reloadOut + "); the client may be ignored until you restart it."
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true, "message": msg, "data": n,
	})
}

func (h *NASHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := PathID(r)
	if !ok {
		WriteErr(w, http.StatusBadRequest, "ID must be greater than 0")
		return
	}
	cur, err := models.GetNASByID(id)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "NAS not found")
		return
	}
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		WriteErr(w, http.StatusBadRequest, "Invalid JSON in request body")
		return
	}
	if len(raw) == 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error",
			"errors": []validation.FieldError{{Field: "body", Message: "At least one field is required"}}})
		return
	}
	var body struct {
		Name        string  `json:"name"`
		IP          string  `json:"ip"`
		Secret      string  `json:"secret"`
		Type        *string `json:"type"`
		Ports       *int    `json:"ports"`
		Community   *string `json:"community"`
		Description *string `json:"description"`
	}
	_ = json.Unmarshal(mustJSON(raw), &body)
	in := validation.NASInput{Name: body.Name, IP: body.IP, Secret: body.Secret}
	if body.Type != nil {
		in.Type = *body.Type
		in.HasType = true
	}
	if body.Ports != nil {
		in.Ports = body.Ports
		in.HasPorts = true
	}
	if body.Community != nil {
		in.Community = *body.Community
		in.HasComm = true
	}
	if body.Description != nil {
		in.Description = *body.Description
		in.HasDesc = true
	}
	if errs := validation.ValidateNASUpdate(&in); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	confName, confIP := cur.Name, cur.IP
	if body.Name != "" {
		confName = body.Name
	}
	if body.IP != "" {
		confIP = body.IP
	}
	if body.Name != "" || body.IP != "" {
		exists, err := models.NASExists(confName, confIP, &id)
		if err != nil {
			WriteInternal(w, "Internal server error", err)
			return
		}
		if exists {
			WriteErr(w, http.StatusConflict, "NAS with this name or IP already exists")
			return
		}
	}
	typ := ""
	if in.HasType {
		typ = in.Type
	}
	community := cur.Community
	if in.HasComm {
		community = in.Community
	}
	desc := cur.Description
	if in.HasDesc {
		desc = in.Description
	}
	ports := 0
	if in.HasPorts && in.Ports != nil {
		ports = *in.Ports
	}
	updated, err := models.UpdateNAS(id, cur, body.Name, body.IP, body.Secret, typ, community, desc, ports)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if updated == nil {
		WriteErr(w, http.StatusNotFound, "NAS not found or no changes made")
		return
	}
	reloadOut, reloadOK := h.reloadRadius()
	msg := "NAS updated successfully. FreeRADIUS restarted to apply the change."
	if !reloadOK {
		msg = "NAS updated, but FreeRADIUS could not be restarted (" + reloadOut + "); the change may not take effect until you restart it."
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": msg, "data": updated,
	})
}

func (h *NASHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := PathID(r)
	if !ok {
		WriteErr(w, http.StatusBadRequest, "ID must be greater than 0")
		return
	}
	cur, err := models.GetNASByID(id)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "NAS not found")
		return
	}
	okDel, err := models.DeleteNAS(id)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if !okDel {
		WriteErr(w, http.StatusNotFound, "NAS not found or already deleted")
		return
	}
	reloadOut, reloadOK := h.reloadRadius()
	msg := "NAS deleted successfully. FreeRADIUS restarted to drop the client."
	if !reloadOK {
		msg = "NAS deleted, but FreeRADIUS could not be restarted (" + reloadOut + "); the client may still be accepted until you restart it."
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": msg,
	})
}

func mustJSON(m map[string]json.RawMessage) []byte {
	b, _ := json.Marshal(m)
	return b
}

var _ = chi.URLParam

// reloadRadius re-reads the FreeRADIUS client list. FreeRADIUS with
// read_clients = yes loads the nas table only at startup, so a freshly added
// NAS stays "unknown client" until the server is restarted. The command comes
// from FREERADIUS_RELOAD_CMD (default "systemctl restart freeradius"); set it
// to just "kill -HUP" if a SIGHUP is enough for your setup, or empty to skip.
// The panel API process must be allowed to run the command (e.g. root).
func (h *NASHandler) reloadRadius() (string, bool) {
	cmd := strings.TrimSpace(h.cfg.RadiusReloadCmd)
	if cmd == "" {
		return "", true
	}
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", true
	}
	out, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)) + " " + err.Error(), false
	}
	return strings.TrimSpace(string(out)), true
}
