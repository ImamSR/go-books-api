package books

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

import "log"

type Handler struct {
	Store Store
}

func NewHandler(s Store) *Handler { return &Handler{Store: s} }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// POST /books
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in Book
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "message": "invalid json"})
		return
	}
	id, err := h.Store.Create(&in)
	if err != nil {
	switch err {
	case ErrInvalidName:
		writeJSON(w, http.StatusBadRequest, map[string]any{"status":"fail","message":"name is required"})
	case ErrReadPageTooBig:
		writeJSON(w, http.StatusBadRequest, map[string]any{"status":"fail","message":"readPage must be <= pageCount"})
	default:
		log.Printf("[books.Create] insert error: %v", err) // <-- tambahkan ini
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status":"error"})
	}
	return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"status": "success",
		"data":   map[string]string{"bookId": id},
	})
}

// GET /books?name=&reading=0|1&finished=0|1
func atoiDef(s string, def int) int {
  if n, err := strconv.Atoi(s); err == nil && n >= 0 { return n }
  return def
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
  q := r.URL.Query()
  name := q.Get("name")

  var readingPtr *bool
  if v := q.Get("reading"); v == "0" || v == "1" { b := v == "1"; readingPtr = &b }
  var finishedPtr *bool
  if v := q.Get("finished"); v == "0" || v == "1" { b := v == "1"; finishedPtr = &b }

  limit := atoiDef(q.Get("limit"), 10)
  if limit > 100 { limit = 100 } // cap
  offset := atoiDef(q.Get("offset"), 0)

  items, total, _ := h.Store.List(Filter{
    Name: name, Reading: readingPtr, Finished: finishedPtr,
    Limit: limit, Offset: offset,
  })

  type light struct{ ID, Name, Publisher string }
  out := make([]light, 0, len(items))
  for _, b := range items {
    out = append(out, light{ID: b.ID, Name: b.Name, Publisher: b.Publisher})
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "status": "success",
    "data":   map[string]any{"books": out},
    "meta":   map[string]any{"limit": limit, "offset": offset, "total": total},
  })
}

// GET /books/{id}
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/books/")
	b, err := h.Store.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "fail", "message": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": map[string]any{"book": b}})
}

// PUT /books/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/books/")
	var in Book
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "message": "invalid json"})
		return
	}
	if err := h.Store.Update(id, in); err != nil {
		switch err {
		case ErrInvalidName:
			writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "message": "name is required"})
		case ErrReadPageTooBig:
			writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "message": "readPage must be <= pageCount"})
		case ErrNotFound:
			writeJSON(w, http.StatusNotFound, map[string]any{"status": "fail", "message": "id not found"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "updated"})
}

// DELETE /books/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/books/")
	if err := h.Store.Delete(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "fail", "message": "id not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "deleted"})
}

// tiny helper (not essential)
func atoi(s string) (int, error) { return strconv.Atoi(s) }
