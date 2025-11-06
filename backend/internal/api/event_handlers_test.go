package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/services"
)

func TestServer_handleGetSwappableEvents(t *testing.T) {
	queries := repository.SetupTestDB(t)
	// dbConn := queries.DB()
	// defer dbConn.Close()

	userRepo := repository.NewUserRepository(queries)
	eventRepo := repository.NewEventRepository(queries)
	swapRepo := repository.NewSwapRequestRepository(queries)
	authService := services.NewAuthService(userRepo, nil, nil) // Mocks
	eventService := services.NewEventService(eventRepo, userRepo, swapRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	server := NewServer(authService, nil, eventService, swapRequestService, nil)

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
	_, err = eventRepo.CreateEvent(context.Background(), db.CreateEventParams{Title: "Event 1", UserID: user1.ID, Status: "SWAPPABLE"})
	if err != nil {
		t.Fatalf("Failed to create event1: %v", err)
	}

	event2, err := eventRepo.CreateEvent(context.Background(), db.CreateEventParams{Title: "Event 2", UserID: user2.ID, Status: "SWAPPABLE"})
	if err != nil {
		t.Fatalf("Failed to create event2: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/swappable-slots", nil)
	ctx := context.WithValue(req.Context(), userIDContextKey, user1.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleGetSwappableEvents)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var events []db.GetSwappableEventsRow
	if err := json.Unmarshal(rr.Body.Bytes(), &events); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	expectedEvent := db.GetSwappableEventsRow{
		ID:        event2.ID,
		Title:     event2.Title,
		StartTime: event2.StartTime,
		EndTime:   event2.EndTime,
		Status:    event2.Status,
		UserID:    event2.UserID,
		CreatedAt: event2.CreatedAt,
		UpdatedAt: event2.UpdatedAt,
		OwnerName: user2.Name,
	}

	// reflect.DeepEqual is not perfect with time.Time, but for this test it should be sufficient
	if !reflect.DeepEqual(events[0], expectedEvent) {
		t.Errorf("Unexpected event data.\nGot:  %+v\nWant: %+v", events[0], expectedEvent)
	}
}

func TestServer_handleGetEventsByUserIDAndStatus(t *testing.T) {
	queries := repository.SetupTestDB(t)
	// dbConn := queries.DB()
	// defer dbConn.Close()

	userRepo := repository.NewUserRepository(queries)
	eventRepo := repository.NewEventRepository(queries)
	swapRepo := repository.NewSwapRequestRepository(queries)
	authService := services.NewAuthService(userRepo, nil, nil) // Mocks
	eventService := services.NewEventService(eventRepo, userRepo, swapRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	server := NewServer(authService, nil, eventService, swapRequestService, nil)

	// Create a user
	user, err := userRepo.CreateUser(context.Background(), db.CreateUserParams{Name: "User One", Email: "user1@test.com", Password: "password"})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create swappable and busy events
	_, err = eventRepo.CreateEvent(context.Background(), db.CreateEventParams{Title: "Swappable Event", UserID: user.ID, Status: "SWAPPABLE"})
	if err != nil {
		t.Fatalf("Failed to create swappable event: %v", err)
	}

	_, err = eventRepo.CreateEvent(context.Background(), db.CreateEventParams{Title: "Busy Event", UserID: user.ID, Status: "BUSY"})
	if err != nil {
		t.Fatalf("Failed to create busy event: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/events/user?status=SWAPPABLE", nil)
	ctx := context.WithValue(req.Context(), userIDContextKey, user.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleGetEventsByUserID)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var events []db.Event
	if err := json.Unmarshal(rr.Body.Bytes(), &events); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	if events[0].Status != "SWAPPABLE" {
		t.Errorf("Expected event status to be SWAPPABLE, got %s", events[0].Status)
	}
}
