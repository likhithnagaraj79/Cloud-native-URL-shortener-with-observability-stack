package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/likhi/url-shortener/internal/models"
	"github.com/likhi/url-shortener/internal/repository"
	"github.com/likhi/url-shortener/internal/service"
)

type URLHandler struct {
	svc    service.URLShortener
	logger *zap.Logger
}

func NewURLHandler(svc service.URLShortener, logger *zap.Logger) *URLHandler {
	return &URLHandler{svc: svc, logger: logger}
}

func (h *URLHandler) Create(c *gin.Context) {
	var req models.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.CreateShortURL(c.Request.Context(), &req, c.Request.UserAgent())
	if err != nil {
		if errors.Is(err, service.ErrCodeTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to create short url", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	original, err := h.svc.Resolve(c.Request.Context(), code)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "short URL not found or expired"})
		return
	}
	if err != nil {
		h.logger.Error("failed to resolve short url", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Redirect(http.StatusMovedPermanently, original)
}

func (h *URLHandler) Stats(c *gin.Context) {
	code := c.Param("code")

	stats, err := h.svc.GetStats(c.Request.Context(), code)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "short URL not found"})
		return
	}
	if err != nil {
		h.logger.Error("failed to get stats", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *URLHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
