package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestHealthCheck(t *testing.T) {
	// Initialize test server
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	_ = e.NewContext(req, rec)

	// Assertions
	// if assert.NoError(t, healthCheck(c)) {
	// 	assert.Equal(t, http.StatusOK, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "ok")
	// }
}
