package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/services"
)

func TestServer_handleGetIncomingSwapRequests(t *testing.T) {
	queries := repository.SetupTestDB(t)

	userRepo := repository.NewUserRepository(queries)
	eventRepo := repository.NewEventRepository(queries)
	swapRepo := repository.NewSwapRequestRepository(queries)
	authService := services.NewAuthService(userRepo, nil, nil) // Mocks
	eventService := services.NewEventService(eventRepo, userRepo, swapRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	server := NewServer(nil, authService, nil, eventService, swapRequestService, nil)

	// Create two users
	user1, err := userRepo.CreateUser(context.Background(), db.CreateUserParams{Name: "User One", Email: "user1@test.com", Password: "password"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := userRepo.CreateUser(context.Background(), db.CreateUserParams{Name: "User Two", Email: "user2@test.com", Password: "password"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create swappable events
	event1, err := eventRepo.CreateEvent(context.Background(), db.CreateEventParams{Title: "Event 1", UserID: user1.ID, Status: "SWAPPABLE"})
	if err != nil {
		t.Fatalf("Failed to create event1: %v", err)
	}

	event2, err := eventRepo.CreateEvent(context.Background(), db.CreateEventParams{Title: "Event 2", UserID: user2.ID, Status: "SWAPPABLE"})
	if err != nil {
		t.Fatalf("Failed to create event2: %v", err)
	}

	// Create a swap request from user1 to user2
	_, err = swapRequestService.CreateSwapRequest(context.Background(), services.CreateSwapRequestInput{
		RequesterUserID: user1.ID,
		ResponderUserID: user2.ID,
		RequesterSlotID: event1.ID,
		ResponderSlotID: event2.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create swap request: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/swap-requests/incoming", nil)
	ctx := context.WithValue(req.Context(), userIDContextKey, user2.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleGetIncomingSwapRequests)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var requests []db.GetIncomingSwapRequestsRow
	if err := json.Unmarshal(rr.Body.Bytes(), &requests); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("Expected 1 incoming request, got %d", len(requests))
	}

	if requests[0].RequesterName != user1.Name {
		t.Errorf("Expected requester name to be %s, got %s", user1.Name, requests[0].RequesterName)
	}
}

func TestServer_handleGetOutgoingSwapRequests(t *testing.T) {
	queries := repository.SetupTestDB(t)

	userRepo := repository.NewUserRepository(queries)
	eventRepo := repository.NewEventRepository(queries)
	swapRepo := repository.NewSwapRequestRepository(queries)
	authService := services.NewAuthService(userRepo, nil, nil) // Mocks
	eventService := services.NewEventService(eventRepo, userRepo, swapRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	server := NewServer(nil, authService, nil, eventService, swapRequestService, nil)

	// Create two users
	user1, err := userRepo.CreateUser(context.Background(), db.CreateUserParams{Name: "User One", Email: "user1@test.com", Password: "password"})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := userRepo.CreateUser(context.Background(), db.CreateUserParams{Name: "User Two", Email: "user2@test.com", Password: "password"})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create swappable events
	event1, err := eventRepo.CreateEvent(context.Background(), db.CreateEventParams{Title: "Event 1", UserID: user1.ID, Status: "SWAPPABLE"})
	if err != nil {
		t.Fatalf("Failed to create event1: %v", err)
	}

	event2, err := eventRepo.CreateEvent(context.Background(), db.CreateEventParams{Title: "Event 2", UserID: user2.ID, Status: "SWAPPABLE"})
	if err != nil {
		t.Fatalf("Failed to create event2: %v", err)
	}

	// Create a swap request from user1 to user2
	_, err = swapRequestService.CreateSwapRequest(context.Background(), services.CreateSwapRequestInput{
		RequesterUserID: user1.ID,
		ResponderUserID: user2.ID,
		RequesterSlotID: event1.ID,
		ResponderSlotID: event2.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create swap request: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/swap-requests/outgoing", nil)
	ctx := context.WithValue(req.Context(), userIDContextKey, user1.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleGetOutgoingSwapRequests)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var requests []db.GetOutgoingSwapRequestsRow
	if err := json.Unmarshal(rr.Body.Bytes(), &requests); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("Expected 1 outgoing request, got %d", len(requests))
	}

	if requests[0].ResponderName != user2.Name {
		t.Errorf("Expected responder name to be %s, got %s", user2.Name, requests[0].ResponderName)
	}
}
