package users

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"github.com/ImamSR/go-books-api/internal/util"
)

import "log"


type Handler struct {
	Repo     Repo
	TokenGen func(sub string, roles []string) (string, error)
}

func NewHandler(r Repo, gen func(string, []string) (string, error)) *Handler {
	return &Handler{Repo: r, TokenGen: gen}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func normalizeEmail(s string) string { return strings.TrimSpace(strings.ToLower(s)) }
func localPart(email string) string {
    if i := strings.IndexByte(email, '@'); i > 0 {
        return email[:i]
    }
    return email
}

// POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "message": "invalid json"})
		return
	}
	in.Email = normalizeEmail(in.Email)
	if in.Email == "" || in.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status":"fail","message":"email & password required"})
		return
	}
	uname := strings.TrimSpace(in.Username)
	if uname == "" {
		uname = localPart(in.Email)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error"})
		return
	}
	if len(in.Roles) == 0 {
		in.Roles = []string{"editor"}
	}
	u := &User{
		ID:       util.RandomID(),
		Email:    in.Email,
		Username: uname,              // <-- set
		Password: string(hash),
		Roles:    in.Roles,
	}
	if err := h.Repo.Create(u); err != nil {
		if err == ErrEmailTaken {
			writeJSON(w, http.StatusConflict, map[string]any{"status":"fail","message":"email already used"})
			return
		}
		log.Printf("[users.Register] create error: %v", err)  // <--- tambahkan ini
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status":"error"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"status": "success"})
}

// POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "message": "invalid json"})
		return
	}
	in.Email = normalizeEmail(in.Email)
	u, err := h.Repo.FindByEmail(in.Email)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"status": "fail", "message": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"status": "fail", "message": "invalid credentials"})
		return
	}
	token, err := h.TokenGen(u.ID, u.Roles)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "success",
		"data":   map[string]any{"accessToken": token},
	})
}