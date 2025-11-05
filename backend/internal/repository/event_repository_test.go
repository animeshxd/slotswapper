package repository

import (
	"context"
	"testing"
	"time"

	"slotswapper/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

func TestEventRepository(t *testing.T) {
	t.Run("CreateEvent", func(t *testing.T) {
		testQueries, user := SetupTestDBWithUser(t)
		eventRepo := NewEventRepository(testQueries)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		arg := db.CreateEventParams{
			Title:     "Test Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		event, err := eventRepo.CreateEvent(context.Background(), arg)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		if event.ID == 0 {
			t.Error("expected event ID to be non-zero")
		}
		if event.Title != arg.Title {
			t.Errorf("expected event title to be %q, got %q", arg.Title, event.Title)
		}
	})

	t.Run("GetEventByID", func(t *testing.T) {
		testQueries, user := SetupTestDBWithUser(t)
		eventRepo := NewEventRepository(testQueries)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		arg := db.CreateEventParams{
			Title:     "Test Event 2",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		createdEvent, err := eventRepo.CreateEvent(context.Background(), arg)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		retrievedEvent, err := eventRepo.GetEventByID(context.Background(), createdEvent.ID)
		if err != nil {
			t.Fatalf("failed to get event by ID: %v", err)
		}

		if retrievedEvent.ID != createdEvent.ID {
			t.Errorf("expected event ID to be %d, got %d", createdEvent.ID, retrievedEvent.ID)
		}
	})

	t.Run("GetEventsByUserID", func(t *testing.T) {
		testQueries, user := SetupTestDBWithUser(t)
		eventRepo := NewEventRepository(testQueries)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		for i := 0; i < 3; i++ {
			arg := db.CreateEventParams{
				Title:     "User Event",
				StartTime: startTime,
				EndTime:   endTime,
				Status:    "BUSY",
				UserID:    user.ID,
			}
			_, err := eventRepo.CreateEvent(context.Background(), arg)
			if err != nil {
				t.Fatalf("failed to create event for user: %v", err)
			}
		}

		events, err := eventRepo.GetEventsByUserID(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("failed to get events by user ID: %v", err)
		}

		if len(events) != 3 {
			t.Errorf("expected 3 events, got %d", len(events))
		}
	})

	t.Run("UpdateEventStatus", func(t *testing.T) {
		testQueries, user := SetupTestDBWithUser(t)
		eventRepo := NewEventRepository(testQueries)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		arg := db.CreateEventParams{
			Title:     "Update Status Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		createdEvent, err := eventRepo.CreateEvent(context.Background(), arg)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		updateArg := db.UpdateEventStatusParams{
			ID:     createdEvent.ID,
			Status: "SWAPPABLE",
		}

		updatedEvent, err := eventRepo.UpdateEventStatus(context.Background(), updateArg)
		if err != nil {
			t.Fatalf("failed to update event status: %v", err)
		}

		if updatedEvent.Status != "SWAPPABLE" {
			t.Errorf("expected event status to be SWAPPABLE, got %s", updatedEvent.Status)
		}
	})

	t.Run("GetSwappableEvents", func(t *testing.T) {
		testQueries, user := SetupTestDBWithUser(t)
		eventRepo := NewEventRepository(testQueries)

		otherUser, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "other user",
			Email:    "other-user@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create other user: %v", err)
		}

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		arg := db.CreateEventParams{
			Title:     "Swappable Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "SWAPPABLE",
			UserID:    otherUser.ID,
		}
		_, err = eventRepo.CreateEvent(context.Background(), arg)
		if err != nil {
			t.Fatalf("failed to create swappable event: %v", err)
		}

		swappableEvents, err := eventRepo.GetSwappableEvents(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("failed to get swappable events: %v", err)
		}

		if len(swappableEvents) != 1 {
			t.Errorf("expected 1 swappable event, got %d", len(swappableEvents))
		}
	})
}
