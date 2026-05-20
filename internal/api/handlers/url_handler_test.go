package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/likhi/url-shortener/internal/api/handlers"
	"github.com/likhi/url-shortener/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRouter(svc *mockSvc) *gin.Engine {
	r := gin.New()
	h := handlers.NewURLHandler(svc, zap.NewNop())
	r.GET("/health", h.Health)
	r.POST("/api/v1/urls", h.Create)
	r.GET("/api/v1/urls/:code/stats", h.Stats)
	r.GET("/:code", h.Redirect)
	return r
}

func do(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- /health ---

func TestHealth(t *testing.T) {
	w := do(newRouter(&mockSvc{}), "GET", "/health", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// --- POST /api/v1/urls ---

func TestCreate_Success(t *testing.T) {
	svc := &mockSvc{createResp: &models.CreateURLResponse{
		ShortCode:   "abc1234",
		ShortURL:    "http://localhost/abc1234",
		OriginalURL: "https://example.com",
		CreatedAt:   time.Now(),
	}}
	w := do(newRouter(svc), "POST", "/api/v1/urls", map[string]string{
		"original_url": "https://example.com",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}
}

func TestCreate_BadRequest(t *testing.T) {
	w := do(newRouter(&mockSvc{}), "POST", "/api/v1/urls", map[string]string{
		"original_url": "not-a-url",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCreate_Conflict(t *testing.T) {
	svc := &mockSvc{createErr: errCodeTaken}
	w := do(newRouter(svc), "POST", "/api/v1/urls", map[string]string{
		"original_url": "https://example.com",
		"custom_code":  "taken",
	})
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}
}

func TestCreate_InternalError(t *testing.T) {
	svc := &mockSvc{createErr: errInternal}
	w := do(newRouter(svc), "POST", "/api/v1/urls", map[string]string{
		"original_url": "https://example.com",
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// --- GET /:code ---

func TestRedirect_Found(t *testing.T) {
	svc := &mockSvc{resolveURL: "https://destination.com"}
	w := do(newRouter(svc), "GET", "/abc1234", nil)
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("status = %d, want 301", w.Code)
	}
	if got := w.Header().Get("Location"); got != "https://destination.com" {
		t.Errorf("Location = %q", got)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	svc := &mockSvc{resolveErr: errNotFound}
	w := do(newRouter(svc), "GET", "/missing", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestRedirect_InternalError(t *testing.T) {
	svc := &mockSvc{resolveErr: errInternal}
	w := do(newRouter(svc), "GET", "/badcode", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// --- GET /api/v1/urls/:code/stats ---

func TestStats_Found(t *testing.T) {
	svc := &mockSvc{statsResp: &models.URLStats{
		ShortCode:   "abc1234",
		OriginalURL: "https://example.com",
		ClickCount:  42,
		CreatedAt:   time.Now(),
	}}
	w := do(newRouter(svc), "GET", "/api/v1/urls/abc1234/stats", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var resp models.URLStats
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ClickCount != 42 {
		t.Errorf("click_count = %d, want 42", resp.ClickCount)
	}
}

func TestStats_NotFound(t *testing.T) {
	svc := &mockSvc{statsErr: errNotFound}
	w := do(newRouter(svc), "GET", "/api/v1/urls/nope/stats", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
