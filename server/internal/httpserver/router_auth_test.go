package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"tavily-proxy/server/internal/db"
	"tavily-proxy/server/internal/services"
)

func TestAdminAuth_LoginLogoutAndProtectedRoute(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	database, err := db.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	master := services.NewMasterKeyService(database, logger)
	if err := master.LoadOrCreate(context.Background()); err != nil {
		t.Fatalf("master key init: %v", err)
	}

	admin, err := services.NewAdminAuthService("admin", "s3cret", time.Hour)
	if err != nil {
		t.Fatalf("new admin auth service: %v", err)
	}

	router := NewRouter(Dependencies{
		MasterKeyService: master,
		AdminAuthService: admin,
	})

	loginBody := []byte(`{"username":"admin","password":"s3cret"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)

	if loginResp.Code != http.StatusOK {
		t.Fatalf("unexpected login status: got %d want %d (body=%q)", loginResp.Code, http.StatusOK, loginResp.Body.String())
	}

	var loginOut struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginOut); err != nil {
		t.Fatalf("login response json: %v (body=%q)", err, loginResp.Body.String())
	}
	if loginOut.Token == "" {
		t.Fatalf("login token should not be empty")
	}

	unauthReq := httptest.NewRequest(http.MethodGet, "/api/keys", nil)
	unauthResp := httptest.NewRecorder()
	router.ServeHTTP(unauthResp, unauthReq)
	if unauthResp.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected unauth status: got %d want %d", unauthResp.Code, http.StatusUnauthorized)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+loginOut.Token)
	logoutResp := httptest.NewRecorder()
	router.ServeHTTP(logoutResp, logoutReq)
	if logoutResp.Code != http.StatusNoContent {
		t.Fatalf("unexpected logout status: got %d want %d", logoutResp.Code, http.StatusNoContent)
	}

	logoutReq2 := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq2.Header.Set("Authorization", "Bearer "+loginOut.Token)
	logoutResp2 := httptest.NewRecorder()
	router.ServeHTTP(logoutResp2, logoutReq2)
	if logoutResp2.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected second logout status: got %d want %d", logoutResp2.Code, http.StatusUnauthorized)
	}
}

func TestAdminAuth_LoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	database, err := db.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	master := services.NewMasterKeyService(database, logger)
	if err := master.LoadOrCreate(context.Background()); err != nil {
		t.Fatalf("master key init: %v", err)
	}

	admin, err := services.NewAdminAuthService("admin", "s3cret", time.Hour)
	if err != nil {
		t.Fatalf("new admin auth service: %v", err)
	}

	router := NewRouter(Dependencies{
		MasterKeyService: master,
		AdminAuthService: admin,
	})

	loginBody := []byte(`{"username":"admin","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthSettings_UpdateCredentialsAndMasterKey(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	database, err := db.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	master := services.NewMasterKeyService(database, logger)
	if err := master.LoadOrCreate(context.Background()); err != nil {
		t.Fatalf("master key init: %v", err)
	}

	admin, err := services.NewAdminAuthService("admin", "s3cret", time.Hour)
	if err != nil {
		t.Fatalf("new admin auth service: %v", err)
	}
	settings := services.NewSettingsService(database)

	router := NewRouter(Dependencies{
		MasterKeyService: master,
		AdminAuthService: admin,
		SettingsService:  settings,
	})

	loginBody := []byte(`{"username":"admin","password":"s3cret"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("unexpected login status: got %d want %d (body=%q)", loginResp.Code, http.StatusOK, loginResp.Body.String())
	}

	var loginOut struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginOut); err != nil {
		t.Fatalf("login response json: %v", err)
	}
	if loginOut.Token == "" {
		t.Fatalf("login token should not be empty")
	}

	updateBody := []byte(`{"master_key":"manually-set-master-key","admin_username":"newadmin","admin_password":"n3w-pass"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/settings/auth", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+loginOut.Token)
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusNoContent {
		t.Fatalf("unexpected update status: got %d want %d (body=%q)", updateResp.Code, http.StatusNoContent, updateResp.Body.String())
	}

	// Existing session must be invalidated after credentials update.
	protectedReq := httptest.NewRequest(http.MethodGet, "/api/settings/auth", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+loginOut.Token)
	protectedResp := httptest.NewRecorder()
	router.ServeHTTP(protectedResp, protectedReq)
	if protectedResp.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected protected status with stale token: got %d want %d", protectedResp.Code, http.StatusUnauthorized)
	}

	oldLoginBody := []byte(`{"username":"admin","password":"s3cret"}`)
	oldLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(oldLoginBody))
	oldLoginReq.Header.Set("Content-Type", "application/json")
	oldLoginResp := httptest.NewRecorder()
	router.ServeHTTP(oldLoginResp, oldLoginReq)
	if oldLoginResp.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected old credential login status: got %d want %d", oldLoginResp.Code, http.StatusUnauthorized)
	}

	newLoginBody := []byte(`{"username":"newadmin","password":"n3w-pass"}`)
	newLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(newLoginBody))
	newLoginReq.Header.Set("Content-Type", "application/json")
	newLoginResp := httptest.NewRecorder()
	router.ServeHTTP(newLoginResp, newLoginReq)
	if newLoginResp.Code != http.StatusOK {
		t.Fatalf("unexpected new credential login status: got %d want %d (body=%q)", newLoginResp.Code, http.StatusOK, newLoginResp.Body.String())
	}

	var newLoginOut struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(newLoginResp.Body.Bytes(), &newLoginOut); err != nil {
		t.Fatalf("new login response json: %v", err)
	}
	if newLoginOut.Token == "" {
		t.Fatalf("new login token should not be empty")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/settings/auth", nil)
	getReq.Header.Set("Authorization", "Bearer "+newLoginOut.Token)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("unexpected get auth settings status: got %d want %d (body=%q)", getResp.Code, http.StatusOK, getResp.Body.String())
	}

	var settingsOut struct {
		MasterKey     string `json:"master_key"`
		AdminUsername string `json:"admin_username"`
	}
	if err := json.Unmarshal(getResp.Body.Bytes(), &settingsOut); err != nil {
		t.Fatalf("get settings response json: %v", err)
	}
	if settingsOut.MasterKey != "manually-set-master-key" {
		t.Fatalf("unexpected master key: got %q", settingsOut.MasterKey)
	}
	if settingsOut.AdminUsername != "newadmin" {
		t.Fatalf("unexpected admin username: got %q", settingsOut.AdminUsername)
	}
}
