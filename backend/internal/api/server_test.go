package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"slotswapper/internal/crypto"
	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/services"

	_ "github.com/mattn/go-sqlite3"
)

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

func signUpAndLogin(t *testing.T, ts *httptest.Server, authService services.AuthService, name, email, password string) (string, *db.User) {
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
	ts, testQueries, jwtManager := setupTestServer(t)
	defer ts.Close()

	authService := services.NewAuthService(repository.NewUserRepository(testQueries), crypto.NewPassword(), jwtManager)
	token, user := signUpAndLogin(t, ts, authService, "Event User", "event@example.com", "eventpassword")

	t.Run("CreateEvent", func(t *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		input := services.CreateEventInput{
			Title:     "My New Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
		}
		body, _ := json.Marshal(input)

		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/events", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var event db.Event
		json.NewDecoder(rr.Body).Decode(&event)
		if event.ID == 0 {
			t.Error("expected event ID to be non-zero")
		}
	})

	t.Run("GetEventByID", func(t *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		createInput := services.CreateEventInput{
			Title:     "Event to Get",
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
			t.Fatalf("failed to create event: %s", createRr.Body.String())
		}
		var createdEvent db.Event
		json.NewDecoder(createRr.Body).Decode(&createdEvent)

		req, _ := http.NewRequest(http.MethodGet, ts.URL+fmt.Sprintf("/api/events/%d", createdEvent.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var retrievedEvent db.Event
		json.NewDecoder(rr.Body).Decode(&retrievedEvent)
		if retrievedEvent.ID != createdEvent.ID {
			t.Errorf("expected event ID %d, got %d", createdEvent.ID, retrievedEvent.ID)
		}
	})

	t.Run("GetEventsByUserID", func(t *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		createInput := services.CreateEventInput{
			Title:     "User Event for List",
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
			t.Fatalf("failed to create event: %s", createRr.Body.String())
		}

		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/events/user", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var events []db.Event
		json.NewDecoder(rr.Body).Decode(&events)
		if len(events) == 0 {
			t.Error("expected at least one event")
		}
		if events[0].UserID != user.ID {
			t.Errorf("expected event to be owned by user %d, got %d", user.ID, events[0].UserID)
		}
	})

	t.Run("UpdateEventStatus", func(t *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		createInput := services.CreateEventInput{
			Title:     "Event to Update",
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
			t.Fatalf("failed to create event: %s", createRr.Body.String())
		}
		var createdEvent db.Event
		json.NewDecoder(createRr.Body).Decode(&createdEvent)

		updateInput := services.UpdateEventStatusInput{
			Status: "SWAPPABLE",
		}
		updateBody, _ := json.Marshal(updateInput)
		updateReq, _ := http.NewRequest(http.MethodPut, ts.URL+fmt.Sprintf("/api/events/%d", createdEvent.ID), bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateRr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(updateRr, updateReq)

		if updateRr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, updateRr.Code, updateRr.Body.String())
		}

		var updatedEvent db.Event
		json.NewDecoder(updateRr.Body).Decode(&updatedEvent)
		if updatedEvent.Status != "SWAPPABLE" {
			t.Errorf("expected event status to be SWAPPABLE, got %s", updatedEvent.Status)
		}
	})

	t.Run("DeleteEvent", func(t *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		createInput := services.CreateEventInput{
			Title:     "Event to Delete",
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
			t.Fatalf("failed to create event: %s", createRr.Body.String())
		}
		var createdEvent db.Event
		json.NewDecoder(createRr.Body).Decode(&createdEvent)

		req, _ := http.NewRequest(http.MethodDelete, ts.URL+fmt.Sprintf("/api/events/%d", createdEvent.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d: %s", http.StatusNoContent, rr.Code, rr.Body.String())
		}
	})
}

func TestSwapAPI(t *testing.T) {
	ts, testQueries, jwtManager := setupTestServer(t)
	defer ts.Close()

	authService := services.NewAuthService(repository.NewUserRepository(testQueries), crypto.NewPassword(), jwtManager)
	token1, user1 := signUpAndLogin(t, ts, authService, "Swap User 1", "swap1@example.com", "swappassword1")
	token2, user2 := signUpAndLogin(t, ts, authService, "Swap User 2", "swap2@example.com", "swappassword2")

	startTime1 := time.Now()
	endTime1 := startTime1.Add(time.Hour)
	eventInput1 := services.CreateEventInput{
		Title:     "User1 Swappable Event",
		StartTime: startTime1,
		EndTime:   endTime1,
		Status:    "SWAPPABLE",
	}
	body1, _ := json.Marshal(eventInput1)
	req1, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/events", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+token1)
	rr1 := httptest.NewRecorder()
	ts.Config.Handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("failed to create event for user1: %s", rr1.Body.String())
	}
	var event1 db.Event
	json.NewDecoder(rr1.Body).Decode(&event1)

	startTime2 := time.Now().Add(2 * time.Hour)
	endTime2 := startTime2.Add(time.Hour)
	eventInput2 := services.CreateEventInput{
		Title:     "User2 Swappable Event",
		StartTime: startTime2,
		EndTime:   endTime2,
		Status:    "SWAPPABLE",
	}
	body2, _ := json.Marshal(eventInput2)
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/events", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token2)
	rr2 := httptest.NewRecorder()
	ts.Config.Handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("failed to create event for user2: %s", rr2.Body.String())
	}
	var event2 db.Event
	json.NewDecoder(rr2.Body).Decode(&event2)

	t.Run("GetSwappableEvents", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/swappable-slots", nil)
		req.Header.Set("Authorization", "Bearer "+token1)

		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var events []db.Event
		json.NewDecoder(rr.Body).Decode(&events)
		if len(events) == 0 {
			t.Error("expected at least one swappable event")
		}
		if events[0].UserID == user1.ID {
			t.Error("expected swappable event not to be owned by the requesting user")
		}
	})

	t.Run("CreateSwapRequest", func(t *testing.T) {
		input := services.CreateSwapRequestInput{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
		}
		body, _ := json.Marshal(input)

		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/swap-request", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token1)

		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var swapRequest db.SwapRequest
		json.NewDecoder(rr.Body).Decode(&swapRequest)
		if swapRequest.ID == 0 {
			t.Error("expected swap request ID to be non-zero")
		}
	})

}
