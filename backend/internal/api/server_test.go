package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"slotswapper/internal/crypto"
	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/services"
)

// Helper function to create a test server and register routes
func setupTestServer(t *testing.T) (*httptest.Server, *db.Queries, crypto.JWT) {
	testQueries := repository.SetupTestDB(t)
	userRepo := repository.NewUserRepository(testQueries)
	eventRepo := repository.NewEventRepository(testQueries)
	swapRepo := repository.NewSwapRequestRepository(testQueries)

	passwordCrypto := crypto.NewPassword()
	jwtSecret := "test-jwt-secret"
	jwtTTL := time.Minute * 10
	jwtManager := crypto.NewJWT(jwtSecret, jwtTTL)

	authService := services.NewAuthService(userRepo, passwordCrypto, jwtManager)
	userCreator := services.NewUserCreator(userRepo, passwordCrypto)
	eventService := services.NewEventService(eventRepo, userRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	server := NewServer(authService, userCreator, eventService, swapRequestService, jwtManager)
	router := http.NewServeMux()
	server.RegisterRoutes(router)

	return httptest.NewServer(router), testQueries, jwtManager
}

// Helper function to sign up a user and return their token
func signUpAndLogin(t *testing.T, ts *httptest.Server, name, email, password string) (string, *db.User) {
	input := services.RegisterUserInput{
		Name:     name,
		Email:    email,
		Password: password,
	}
	body, _ := json.Marshal(input)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ts.Config.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("failed to sign up user: %s", rr.Body.String())
	}

	var response struct {
		User  db.User `json:"user"`
		Token string  `json:"token"`
	}
	json.NewDecoder(rr.Body).Decode(&response)

	return response.Token, &response.User
}

func TestAuthAPI(t *testing.T) {
	ts, _, _ := setupTestServer(t)
	defer ts.Close()

	t.Run("SignUp", func(t *testing.T) {
		input := services.RegisterUserInput{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(input)

		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var response map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&response)

		if response["user"] == nil {
			t.Error("expected user in response")
		}
		if response["token"] == nil {
			t.Error("expected token in response")
		}
	})

	t.Run("Login", func(t *testing.T) {
		// First, sign up a user
		signUpInput := services.RegisterUserInput{
			Name:     "Login User",
			Email:    "login@example.com",
			Password: "loginpassword",
		}
		signUpBody, _ := json.Marshal(signUpInput)
		signUpReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/signup", bytes.NewBuffer(signUpBody))
		signUpReq.Header.Set("Content-Type", "application/json")
		signUpRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(signUpRr, signUpReq)
		if signUpRr.Code != http.StatusOK {
			t.Fatalf("failed to sign up user: %s", signUpRr.Body.String())
		}

		// Then, attempt to log in
		loginInput := services.LoginInput{
			Email:    "login@example.com",
			Password: "loginpassword",
		}
		loginBody, _ := json.Marshal(loginInput)
		loginReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/login", bytes.NewBuffer(loginBody))
		loginReq.Header.Set("Content-Type", "application/json")

		loginRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(loginRr, loginReq)

		if loginRr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, loginRr.Code, loginRr.Body.String())
		}

		var response map[string]interface{}
		json.NewDecoder(loginRr.Body).Decode(&response)

		if response["user"] == nil {
			t.Error("expected user in response")
		}
		if response["token"] == nil {
			t.Error("expected token in response")
		}
	})
}

func TestEventAPI(t *testing.T) {
	ts, _, _ := setupTestServer(t)
	defer ts.Close()

	// Sign up a user to get a token
	token, user := signUpAndLogin(t, ts, "Event User", "event@example.com", "eventpassword")

	t.Run("Event CRUD", func(t *testing.T) {
		// 1. Create Event
		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		createInput := services.CreateEventInput{
			Title:     "My New Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
		}
		createBody, _ := json.Marshal(createInput)
		createReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/events", bytes.NewBuffer(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token)
		createRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(createRr, createReq)

		if createRr.Code != http.StatusOK {
			t.Fatalf("CreateEvent: expected status %d, got %d: %s", http.StatusOK, createRr.Code, createRr.Body.String())
		}
		var createdEvent db.Event
		json.NewDecoder(createRr.Body).Decode(&createdEvent)
		if createdEvent.ID == 0 {
			t.Fatal("CreateEvent: expected event ID to be non-zero")
		}

		// 2. Get Event By ID
		getReq, _ := http.NewRequest(http.MethodGet, ts.URL+fmt.Sprintf("/api/events/%d", createdEvent.ID), nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(getRr, getReq)

		if getRr.Code != http.StatusOK {
			t.Fatalf("GetEventByID: expected status %d, got %d: %s", http.StatusOK, getRr.Code, getRr.Body.String())
		}
		var retrievedEvent db.Event
		json.NewDecoder(getRr.Body).Decode(&retrievedEvent)
		if retrievedEvent.ID != createdEvent.ID {
			t.Errorf("GetEventByID: expected event ID %d, got %d", createdEvent.ID, retrievedEvent.ID)
		}

		// 3. Get Events By User ID
		listReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/events/user", nil)
		listReq.Header.Set("Authorization", "Bearer "+token)
		listRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(listRr, listReq)

		if listRr.Code != http.StatusOK {
			t.Fatalf("GetEventsByUserID: expected status %d, got %d: %s", http.StatusOK, listRr.Code, listRr.Body.String())
		}
		var events []db.Event
		json.NewDecoder(listRr.Body).Decode(&events)
		if len(events) == 0 {
			t.Fatal("GetEventsByUserID: expected at least one event")
		}
		if events[0].UserID != user.ID {
			t.Errorf("GetEventsByUserID: expected event to be owned by user %d, got %d", user.ID, events[0].UserID)
		}

		// 4. Update Event Status
		updateInput := services.UpdateEventStatusInput{Status: "SWAPPABLE"}
		updateBody, _ := json.Marshal(updateInput)
		updateReq, _ := http.NewRequest(http.MethodPut, ts.URL+fmt.Sprintf("/api/events/%d", createdEvent.ID), bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(updateRr, updateReq)

		if updateRr.Code != http.StatusOK {
			t.Fatalf("UpdateEventStatus: expected status %d, got %d: %s", http.StatusOK, updateRr.Code, updateRr.Body.String())
		}
		var updatedEvent db.Event
		json.NewDecoder(updateRr.Body).Decode(&updatedEvent)
		if updatedEvent.Status != "SWAPPABLE" {
			t.Errorf("UpdateEventStatus: expected status SWAPPABLE, got %s", updatedEvent.Status)
		}

		// 5. Delete Event
		deleteReq, _ := http.NewRequest(http.MethodDelete, ts.URL+fmt.Sprintf("/api/events/%d", createdEvent.ID), nil)
		deleteReq.Header.Set("Authorization", "Bearer "+token)
		deleteRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(deleteRr, deleteReq)

		if deleteRr.Code != http.StatusNoContent {
			t.Fatalf("DeleteEvent: expected status %d, got %d: %s", http.StatusNoContent, deleteRr.Code, deleteRr.Body.String())
		}
	})
}

func TestSwapAPI(t *testing.T) {
	ts, _, _ := setupTestServer(t)
	defer ts.Close()

	// Sign up two users
	token1, user1 := signUpAndLogin(t, ts, "Swap User 1", "swap1@example.com", "swappassword1")
	token2, user2 := signUpAndLogin(t, ts, "Swap User 2", "swap2@example.com", "swappassword2")

	// Create swappable events for both users
	createEvent := func(token string) db.Event {
		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		input := services.CreateEventInput{
			Title:     "Swappable Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "SWAPPABLE",
		}
		body, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/events", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("failed to create event: %s", rr.Body.String())
		}
		var event db.Event
		json.NewDecoder(rr.Body).Decode(&event)
		return event
	}

	event1 := createEvent(token1)
	event2 := createEvent(token2)

	t.Run("GetSwappableEvents", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/swappable-slots", nil)
		req.Header.Set("Authorization", "Bearer "+token1)
		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("GetSwappableEvents: expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		var events []db.Event
		json.NewDecoder(rr.Body).Decode(&events)
		if len(events) == 0 {
			t.Fatal("GetSwappableEvents: expected at least one swappable event")
		}
		if events[0].UserID == user1.ID {
			t.Error("GetSwappableEvents: expected swappable event not to be owned by the requesting user")
		}
	})

	t.Run("Full Swap Flow", func(t *testing.T) {
		// 1. Create Swap Request
		createReqInput := services.CreateSwapRequestInput{
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
		}
		createReqBody, _ := json.Marshal(createReqInput)
		createReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/swap-request", bytes.NewBuffer(createReqBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token1)
		createRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(createRr, createReq)

		if createRr.Code != http.StatusOK {
			t.Fatalf("CreateSwapRequest: expected status %d, got %d: %s", http.StatusOK, createRr.Code, createRr.Body.String())
		}
		var createdSwapRequest db.SwapRequest
		json.NewDecoder(createRr.Body).Decode(&createdSwapRequest)
		if createdSwapRequest.ID == 0 {
			t.Fatal("CreateSwapRequest: expected swap request ID to be non-zero")
		}

		// 2. Accept Swap Request
		acceptInput := map[string]string{"status": "ACCEPTED"}
		acceptBody, _ := json.Marshal(acceptInput)
		acceptReq, _ := http.NewRequest(http.MethodPost, ts.URL+fmt.Sprintf("/api/swap-response/%d", createdSwapRequest.ID), bytes.NewBuffer(acceptBody))
		acceptReq.Header.Set("Content-Type", "application/json")
		acceptReq.Header.Set("Authorization", "Bearer "+token2) // Responder accepts
		acceptRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(acceptRr, acceptReq)

		if acceptRr.Code != http.StatusOK {
			t.Fatalf("AcceptSwapRequest: expected status %d, got %d: %s", http.StatusOK, acceptRr.Code, acceptRr.Body.String())
		}
		var acceptedSwapRequest db.SwapRequest
		json.NewDecoder(acceptRr.Body).Decode(&acceptedSwapRequest)
		if acceptedSwapRequest.Status != "ACCEPTED" {
			t.Errorf("AcceptSwapRequest: expected status ACCEPTED, got %s", acceptedSwapRequest.Status)
		}
	})
}