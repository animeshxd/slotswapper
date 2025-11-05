package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"slotswapper/internal/api"
	"slotswapper/internal/crypto"
	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/services"
)

func main() {
	dbPath := "./db/slotswapper.db"
	dbConn, err := sql.Open("sqlite3", dbPath)
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
	userCreator := services.NewUserCreator(userRepo, passwordCrypto)
	eventService := services.NewEventService(eventRepo, userRepo)
	swapRequestService := services.NewSwapRequestService(swapRepo, eventRepo, userRepo)

	server := api.NewServer(authService, userCreator, eventService, swapRequestService, jwtManager)

	router := http.NewServeMux()
	server.RegisterRoutes(router)

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
