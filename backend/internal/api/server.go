package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"slotswapper/internal/crypto"
	"slotswapper/internal/services"

	"github.com/go-playground/validator/v10"
)

type Server struct {
	authService        services.AuthService
	userCreator        services.UserCreator
	eventService       services.EventService
	swapRequestService services.SwapRequestService
	validator          *validator.Validate
	jwtManager         crypto.JWT
}

func NewServer(authService services.AuthService, userCreator services.UserCreator, eventService services.EventService, swapRequestService services.SwapRequestService, jwtManager crypto.JWT) *Server {
	return &Server{
		authService:        authService,
		userCreator:        userCreator,
		eventService:       eventService,
		swapRequestService: swapRequestService,
		validator:          validator.New(),
		jwtManager:         jwtManager,
	}
}

func (s *Server) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/health", s.healthCheck)

	router.HandleFunc("/api/signup", s.handleSignUp)
	router.HandleFunc("/api/login", s.handleLogin)

	protected := http.NewServeMux()

	protected.HandleFunc("/api/events", s.handleCreateEvent)
	protected.HandleFunc("/api/events/user", s.handleGetEventsByUserID)
	protected.HandleFunc("/api/events/", s.handleEventByID)

	protected.HandleFunc("/api/swappable-slots", s.handleGetSwappableEvents)
	protected.HandleFunc("/api/swap-request", s.handleCreateSwapRequest)
	protected.HandleFunc("/api/swap-response/", s.handleUpdateSwapRequestStatus)

	router.Handle("/api/", AuthMiddleware(s.jwtManager)(protected))
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "OK")
}

func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input services.RegisterUserInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, token, err := s.authService.Register(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"user": user, "token": token})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input services.LoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, token, err := s.authService.Login(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"user": user, "token": token})
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input services.CreateEventInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.UserID = userID // Set user ID from authenticated context

	event, err := s.eventService.CreateEvent(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (s *Server) handleEventByID(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 4 || pathSegments[3] == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}
	eventID, err := strconv.ParseInt(pathSegments[3], 10, 64)
	if err != nil {
		http.Error(w, "Invalid Event ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		event, err := s.eventService.GetEventByID(r.Context(), eventID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(event)
	case http.MethodPut:
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var input services.UpdateEventStatusInput
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		input.ID = eventID
		input.UserID = userID

		updatedEvent, err := s.eventService.UpdateEventStatus(r.Context(), input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedEvent)
	case http.MethodDelete:
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err := s.eventService.DeleteEvent(r.Context(), eventID, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGetEventsByUserID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	events, err := s.eventService.GetEventsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) handleGetSwappableEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	events, err := s.eventService.GetSwappableEvents(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) handleCreateSwapRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input services.CreateSwapRequestInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.RequesterUserID = userID // Set requester ID from authenticated context

	swapRequest, err := s.swapRequestService.CreateSwapRequest(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(swapRequest)
}

func (s *Server) handleUpdateSwapRequestStatus(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 4 || pathSegments[3] == "" {
		http.Error(w, "Swap Request ID is required", http.StatusBadRequest)
		return
	}
	swapRequestID, err := strconv.ParseInt(pathSegments[3], 10, 64)
	if err != nil {
		http.Error(w, "Invalid Swap Request ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input services.UpdateSwapRequestStatusInput
	var status struct {
		Status string `json:"status"`
	}
	err = json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input.ID = swapRequestID
	input.Status = status.Status
	input.UserID = userID

	updatedSwapRequest, err := s.swapRequestService.UpdateSwapRequestStatus(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSwapRequest)
}
