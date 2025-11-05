package api

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"slotswapper/internal/crypto"
	"slotswapper/internal/services"
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
	router.HandleFunc("GET /health", s.healthCheck)

	// Auth routes
	router.HandleFunc("POST /api/signup", s.handleSignUp)
	router.HandleFunc("POST /api/login", s.handleLogin)

	// Protected routes
	// Event routes
	router.Handle("POST /api/events", AuthMiddleware(s.jwtManager)(http.HandlerFunc(s.handleCreateEvent)))
	router.Handle("GET /api/events/user", AuthMiddleware(s.jwtManager)(http.HandlerFunc(s.handleGetEventsByUserID)))
	router.Handle("GET /api/events/{id}", AuthMiddleware(s.jwtManager)(http.HandlerFunc(s.handleGetEventByID)))
	router.Handle("PUT /api/events/{id}", AuthMiddleware(s.jwtManager)(http.HandlerFunc(s.handleUpdateEventStatus)))
	router.Handle("DELETE /api/events/{id}", AuthMiddleware(s.jwtManager)(http.HandlerFunc(s.handleDeleteEvent)))

	// Swap routes
	router.Handle("GET /api/swappable-slots", AuthMiddleware(s.jwtManager)(http.HandlerFunc(s.handleGetSwappableEvents)))
	router.Handle("POST /api/swap-request", AuthMiddleware(s.jwtManager)(http.HandlerFunc(s.handleCreateSwapRequest)))
	router.Handle("POST /api/swap-response/{id}", AuthMiddleware(s.jwtManager)(http.HandlerFunc(s.handleUpdateSwapRequestStatus)))
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}