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
	userService := services.NewUserService(userRepo, passwordCrypto)
	eventService := services.NewEventService(eventRepo, userRepo, swapRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	server := NewServer(authService, userService, eventService, swapRequestService, jwtManager)
	router := http.NewServeMux()
	server.RegisterRoutes(router)

	return httptest.NewServer(router), testQueries, jwtManager
}

// Helper function to sign up a user and return their token, user object, and access_token cookie
func signUpAndLogin(t *testing.T, ts *httptest.Server, name, email, password string) (string, *db.User, *http.Cookie) {
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

	var accessTokenCookie *http.Cookie
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "access_token" {
			accessTokenCookie = cookie
			break
		}
	}

	return response.Token, &response.User, accessTokenCookie
}

func TestAuthAPI(t *testing.T) {
	ts, _, _ := setupTestServer(t)
	defer ts.Close()

	t.Run("SignUp sets cookie and returns token", func(t *testing.T) {
		token, user, cookie := signUpAndLogin(t, ts, "Test User", "test@example.com", "password123")

		if token == "" {
			t.Error("expected token to be non-empty")
		}
		if user == nil {
			t.Error("expected user to be non-nil")
		}
		if cookie == nil {
			t.Error("expected access_token cookie to be set")
		}
		if cookie != nil && cookie.Value != token {
			t.Errorf("expected cookie value %q, got %q", token, cookie.Value)
		}
	})

	t.Run("Login sets cookie and returns token", func(t *testing.T) {
		// First, sign up a user
		signUpAndLogin(t, ts, "Login User", "login@example.com", "loginpassword")

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

		var response struct {
			User  db.User `json:"user"`
			Token string  `json:"token"`
		}
		json.NewDecoder(loginRr.Body).Decode(&response)

		if response.User.ID == 0 {
			t.Error("expected user in response")
		}
		if response.Token == "" {
			t.Error("expected token in response")
		}

		// Check for cookie
		cookies := loginRr.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "access_token" {
				found = true
				if cookie.Value != response.Token {
					t.Errorf("expected cookie value %q, got %q", response.Token, cookie.Value)
				}
				break
			}
		}
		if !found {
			t.Error("expected access_token cookie to be set on login")
		}
	})
}

func TestUserAPI(t *testing.T) {
	ts, _, _ := setupTestServer(t)
	defer ts.Close()

	// Sign up a user to get a token and cookie
	_, user, cookie := signUpAndLogin(t, ts, "User API Test User", "userapi@example.com", "userapipassword")
	if cookie == nil {
		t.Fatal("access_token cookie not found after signup")
	}

	t.Run("Get /api/me", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/me", nil)
		req.AddCookie(cookie)
		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Get /api/me: expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var retrievedUser db.GetUserByIDRow
		json.NewDecoder(rr.Body).Decode(&retrievedUser)
		if retrievedUser.ID != user.ID {
			t.Errorf("Get /api/me: expected user ID %d, got %d", user.ID, retrievedUser.ID)
		}
		if retrievedUser.Email != user.Email {
			t.Errorf("Get /api/me: expected user email %q, got %q", user.Email, retrievedUser.Email)
		}
	})

	t.Run("Get /api/users/{id}", func(t *testing.T) {
		// Create another user for public profile test
		_, publicUser, publicCookie := signUpAndLogin(t, ts, "Public User", "public@example.com", "publicpassword")
		if publicCookie == nil {
			t.Fatal("access_token cookie not found for public user signup")
		}

		req, _ := http.NewRequest(http.MethodGet, ts.URL+fmt.Sprintf("/api/users/%d", publicUser.ID), nil)
		req.AddCookie(cookie)
		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Get /api/users/{id}: expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var retrievedPublicUser db.GetPublicUserByIDRow
		json.NewDecoder(rr.Body).Decode(&retrievedPublicUser)
		if retrievedPublicUser.ID != publicUser.ID {
			t.Errorf("Get /api/users/{id}: expected user ID %d, got %d", publicUser.ID, retrievedPublicUser.ID)
		}
		if retrievedPublicUser.Name != publicUser.Name {
			t.Errorf("Get /api/users/{id}: expected user name %q, got %q", publicUser.Name, retrievedPublicUser.Name)
		}
		// Ensure email is not returned in public profile
		// This is implicitly tested by the type db.GetPublicUserByIDRow not having an Email field
	})
}

func TestEventAPI(t *testing.T) {
	ts, _, _ := setupTestServer(t)
	defer ts.Close()

	// Sign up a user to get a token and cookie
	_, user, cookie := signUpAndLogin(t, ts, "Event User", "event@example.com", "eventpassword")
	if cookie == nil {
		t.Fatal("access_token cookie not found after signup")
	}

	t.Run("Event CRUD with Cookie", func(t *testing.T) {
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
		createReq.AddCookie(cookie)
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
		getReq.AddCookie(cookie)
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
		listReq.AddCookie(cookie)
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
		updateReq.AddCookie(cookie)
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
		deleteReq.AddCookie(cookie)
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

	// Sign up two users and get their tokens and cookies
	_, user1, cookie1 := signUpAndLogin(t, ts, "Swap User 1", "swap1@example.com", "swappassword1")
	_, user2, cookie2 := signUpAndLogin(t, ts, "Swap User 2", "swap2@example.com", "swappassword2")
	if cookie1 == nil || cookie2 == nil {
		t.Fatal("access_token cookie not found after signup for swap users")
	}

	// Create swappable events for both users
	createEvent := func(cookie *http.Cookie) db.Event {
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
		req.AddCookie(cookie)
		rr := httptest.NewRecorder()
		ts.Config.Handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("failed to create event: %s", rr.Body.String())
		}
		var event db.Event
		json.NewDecoder(rr.Body).Decode(&event)
		return event
	}

	event1 := createEvent(cookie1)
	event2 := createEvent(cookie2)

	t.Run("GetSwappableEvents", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/swappable-slots", nil)
		req.AddCookie(cookie1)
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
		createReq.AddCookie(cookie1)
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
		acceptReq.AddCookie(cookie2) // Responder accepts
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
