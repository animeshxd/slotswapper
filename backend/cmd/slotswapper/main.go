package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"

	"slotswapper/internal/api"
	"slotswapper/internal/crypto"
	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/services"
)

func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	dbPath := flag.String("db", "db/slotswapper.db", "path to database file")
	port := os.Getenv("PORT")
	flag.Parse()
	config, err := api.LoadConfig(*configPath)

	dbConn, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer dbConn.Close()

	migrations, err := os.ReadFile("./db/migrations/001_initial_schema.sql")
	if err != nil {
		log.Fatalf("failed to read migrations: %v", err)
	}

	_, err = dbConn.Exec(string(migrations))
	if err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	queries := db.New(dbConn)

	passwordCrypto := crypto.NewPassword()
	jwtManager := crypto.NewJWT("supersecretjwtkey", time.Hour*24) // TODO: Move secret to config

	userRepo := repository.NewUserRepository(queries)
	eventRepo := repository.NewEventRepository(queries)
	swapRepo := repository.NewSwapRequestRepository(queries)

	authService := services.NewAuthService(userRepo, passwordCrypto, jwtManager)
	userService := services.NewUserService(userRepo, passwordCrypto)
	eventService := services.NewEventService(eventRepo, userRepo, swapRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	server := api.NewServer(config, authService, userService, eventService, swapRequestService, jwtManager)

	router := http.NewServeMux()
	server.RegisterRoutes(router)

	// Setup CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true,
	})

	handler := c.Handler(router)

	Addr := ":8080"
	if config != nil && config.Addr != "" {
		Addr = config.Addr
	}
	if port != "" {
		Addr = ":" + port
	}
	log.Printf("Server starting on %s\n", Addr)
	log.Fatal(http.ListenAndServe(Addr, handler))
}
