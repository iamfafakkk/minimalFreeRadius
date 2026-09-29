package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/models"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/validation"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler { return &UserHandler{} }

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := QueryPageLimit(r)
	search := r.URL.Query().Get("search")
	profile := r.URL.Query().Get("profile")
	items, err := models.GetAllUsers(search, profile)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	paged, total, pages := Paginate(items, page, limit)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Users retrieved successfully",
		"data": paged, "count": len(paged),
		"pagination": map[string]interface{}{"page": page, "limit": limit, "total": total, "pages": pages},
	})
}

func (h *UserHandler) Stats(w http.ResponseWriter, r *http.Request) {
	c, err := models.CountUsers()
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "User statistics retrieved successfully",
		"data": map[string]interface{}{"total_users": c},
	})
}

func (h *UserHandler) Profiles(w http.ResponseWriter, r *http.Request) {
	names, err := models.ListProfiles()
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Profiles retrieved successfully", "data": names, "count": len(names),
	})
}

func (h *UserHandler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := PathParam(r, "username")
	if errs := validation.ValidateUsernameParam(username); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	u, err := models.GetUserByUsername(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if u == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "User retrieved successfully", "data": u})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := PathID(r)
	if !ok {
		WriteErr(w, http.StatusBadRequest, "ID must be greater than 0")
		return
	}
	u, err := models.GetUserByID(id)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if u == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "User retrieved successfully", "data": u})
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		User     string `json:"user"`
		Username string `json:"username"`
		Password string `json:"password"`
		Profile  string `json:"profile"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteErr(w, http.StatusBadRequest, "Invalid JSON in request body")
		return
	}
	username := body.User
	if username == "" {
		username = body.Username
	}
	if errs := validation.ValidateUserCreate(username, body.Password); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	exists, err := models.UserExists(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if exists {
		WriteErr(w, http.StatusConflict, "User already exists")
		return
	}
	u, err := models.CreateUser(username, body.Password, body.Profile)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{"success": true, "message": "User created successfully", "data": u})
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	username := PathParam(r, "username")
	if errs := validation.ValidateUsernameParam(username); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	cur, err := models.GetUserByUsername(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
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
		Password *string `json:"password"`
		Profile  *string `json:"profile"`
	}
	_ = json.Unmarshal(mustJSON(raw), &body)
	pw, hasPw := "", false
	if body.Password != nil {
		pw, hasPw = *body.Password, true
	}
	prof, hasProf := "", false
	if body.Profile != nil {
		prof, hasProf = *body.Profile, true
	}
	if errs := validation.ValidateUserUpdate(pw, hasPw); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	updated, err := models.UpdateUser(username, pw, prof, hasPw, hasProf)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if updated == nil {
		WriteErr(w, http.StatusNotFound, "User not found or no changes made")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "User updated successfully", "data": updated})
}

func (h *UserHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	id, ok := PathID(r)
	if !ok {
		WriteErr(w, http.StatusBadRequest, "ID must be greater than 0")
		return
	}
	cur, err := models.GetUserByID(id)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
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
		Password *string `json:"password"`
		Profile  *string `json:"profile"`
	}
	_ = json.Unmarshal(mustJSON(raw), &body)
	pw, hasPw := "", false
	if body.Password != nil {
		pw, hasPw = *body.Password, true
	}
	prof, hasProf := "", false
	if body.Profile != nil {
		prof, hasProf = *body.Profile, true
	}
	if errs := validation.ValidateUserUpdate(pw, hasPw); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	updated, err := models.UpdateUserByID(id, pw, prof, hasPw, hasProf)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if updated == nil {
		WriteErr(w, http.StatusNotFound, "User not found or no changes made")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "User updated successfully", "data": updated})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	username := PathParam(r, "username")
	if errs := validation.ValidateUsernameParam(username); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	cur, err := models.GetUserByUsername(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
		return
	}
	okDel, err := models.DeleteUser(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if !okDel {
		WriteErr(w, http.StatusNotFound, "User not found or already deleted")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "User deleted successfully"})
}

func (h *UserHandler) Attributes(w http.ResponseWriter, r *http.Request) {
	username := PathParam(r, "username")
	if errs := validation.ValidateUsernameParam(username); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	cur, err := models.GetUserByUsername(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
		return
	}
	attrs, err := models.GetUserAttributes(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if attrs == nil {
		attrs = []models.Attribute{}
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "User attributes retrieved successfully",
		"data": map[string]interface{}{"username": username, "attributes": attrs},
	})
}

func (h *UserHandler) ReplyAttributes(w http.ResponseWriter, r *http.Request) {
	username := PathParam(r, "username")
	if errs := validation.ValidateUsernameParam(username); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	cur, err := models.GetUserByUsername(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
		return
	}
	attrs, err := models.GetUserReplyAttributes(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if attrs == nil {
		attrs = []models.Attribute{}
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "User reply attributes retrieved successfully",
		"data": map[string]interface{}{"username": username, "reply_attributes": attrs},
	})
}

func (h *UserHandler) AddAttribute(w http.ResponseWriter, r *http.Request) {
	username := PathParam(r, "username")
	if errs := validation.ValidateUsernameParam(username); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	cur, err := models.GetUserByUsername(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
		return
	}
	var body struct {
		Attribute string `json:"attribute"`
		Op        string `json:"op"`
		Value     string `json:"value"`
		Table     string `json:"table"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteErr(w, http.StatusBadRequest, "Invalid JSON in request body")
		return
	}
	if body.Attribute == "" || body.Value == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error",
			"errors": []validation.FieldError{{Field: "attribute", Message: "attribute and value are required"}}})
		return
	}
	table := body.Table
	if table == "" {
		table = "radcheck"
	}
	if table != "radcheck" && table != "radreply" {
		WriteErr(w, http.StatusBadRequest, "Table must be either \"radcheck\" or \"radreply\"")
		return
	}
	op := body.Op
	if op == "" {
		op = "="
	}
	id, err := models.AddAttribute(username, body.Attribute, op, body.Value, table)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true, "message": "Attribute added successfully",
		"data": map[string]interface{}{"id": id, "username": username, "attribute": body.Attribute, "op": op, "value": body.Value, "table": table},
	})
}

func (h *UserHandler) RemoveAttribute(w http.ResponseWriter, r *http.Request) {
	username := PathParam(r, "username")
	if errs := validation.ValidateUsernameParam(username); len(errs) > 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Validation error", "errors": errs})
		return
	}
	cur, err := models.GetUserByUsername(username)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if cur == nil {
		WriteErr(w, http.StatusNotFound, "User not found")
		return
	}
	var body struct {
		Attribute string `json:"attribute"`
		Table     string `json:"table"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	table := body.Table
	if table == "" {
		table = "radcheck"
	}
	if table != "radcheck" && table != "radreply" {
		WriteErr(w, http.StatusBadRequest, "Table must be either \"radcheck\" or \"radreply\"")
		return
	}
	if body.Attribute == "" {
		WriteErr(w, http.StatusBadRequest, "attribute is required")
		return
	}
	removed, err := models.RemoveAttribute(username, body.Attribute, table)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	if !removed {
		WriteErr(w, http.StatusNotFound, "Attribute not found")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Attribute removed successfully"})
}
