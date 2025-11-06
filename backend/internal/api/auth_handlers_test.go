package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"slotswapper/internal/crypto"
	"slotswapper/internal/repository"
	"slotswapper/internal/services"
)

func TestServer_handleSignUp_DuplicateEmail(t *testing.T) {
	queries := repository.SetupTestDB(t)

	userRepo := repository.NewUserRepository(queries)
	eventRepo := repository.NewEventRepository(queries)
	swapRepo := repository.NewSwapRequestRepository(queries)
    passwordCrypto := crypto.NewPassword()
    jwtManager := crypto.NewJWT("test-secret", time.Minute)
	authService := services.NewAuthService(userRepo, passwordCrypto, jwtManager)
	eventService := services.NewEventService(eventRepo, userRepo, swapRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	server := NewServer(authService, nil, eventService, swapRequestService, nil)

	// First registration should succeed
	input := services.RegisterUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleSignUp)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("first signup returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Second registration with the same email should fail
	req = httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(jsonBody))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("second signup returned wrong status code: got %v want %v", status, http.StatusConflict)
	}
}