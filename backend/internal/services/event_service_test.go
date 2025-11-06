package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"slotswapper/internal/db"
	"slotswapper/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func TestEventService(t *testing.T) {
	t.Run("CreateEvent", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		input := CreateEventInput{
			Title:     "Service Test Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		event, err := eventService.CreateEvent(context.Background(), input)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		if event == nil {
			t.Error("expected event to be non-nil")
		}

		if event.ID == 0 {
			t.Error("expected event ID to be non-zero")
		}
		if event.Title != input.Title {
			t.Errorf("expected event title to be %q, got %q", input.Title, event.Title)
		}
	})

	t.Run("CreateEvent_ValidationErrors", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		testCases := []struct {
			name  string
			input CreateEventInput
		}{
			{
				name: "missing title",
				input: CreateEventInput{
					StartTime: startTime,
					EndTime:   endTime,
					Status:    "BUSY",
					UserID:    user.ID,
				},
			},
			{
				name: "end time before start time",
				input: CreateEventInput{
					Title:     "Invalid Time Event",
					StartTime: endTime,
					EndTime:   startTime,
					Status:    "BUSY",
					UserID:    user.ID,
				},
			},
			{
				name: "invalid status",
				input: CreateEventInput{
					Title:     "Invalid Status Event",
					StartTime: startTime,
					EndTime:   endTime,
					Status:    "INVALID",
					UserID:    user.ID,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := eventService.CreateEvent(context.Background(), tc.input)
				if err == nil {
					t.Fatal("expected a validation error, got nil")
				}
			})
		}
	})

	t.Run("GetEventByID", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		input := CreateEventInput{
			Title:     "Service Test Event 2",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		createdEvent, err := eventService.CreateEvent(context.Background(), input)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		retrievedEvent, err := eventService.GetEventByID(context.Background(), createdEvent.ID)
		if err != nil {
			t.Fatalf("failed to get event by ID: %v", err)
		}

		if retrievedEvent == nil {
			t.Error("expected retrieved event to be non-nil")
		}

		if retrievedEvent.ID != createdEvent.ID {
			t.Errorf("expected event ID to be %d, got %d", createdEvent.ID, retrievedEvent.ID)
		}
	})

	t.Run("GetEventsByUserID", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		for i := 0; i < 3; i++ {
			input := CreateEventInput{
				Title:     "Service User Event",
				StartTime: startTime,
				EndTime:   endTime,
				Status:    "BUSY",
				UserID:    user.ID,
			}
			_, err := eventService.CreateEvent(context.Background(), input)
			if err != nil {
				t.Fatalf("failed to create event for user: %v", err)
			}
		}

		events, err := eventService.GetEventsByUserID(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("failed to get events by user ID: %v", err)
		}

		if len(events) != 3 {
			t.Errorf("expected 3 events, got %d", len(events))
		}
	})

	t.Run("UpdateEventStatus", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		createInput := CreateEventInput{
			Title:     "Service Update Status Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		createdEvent, err := eventService.CreateEvent(context.Background(), createInput)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		updateInput := UpdateEventStatusInput{
			ID:     createdEvent.ID,
			Status: "SWAPPABLE",
			UserID: user.ID,
		}

		updatedEvent, err := eventService.UpdateEventStatus(context.Background(), updateInput)
		if err != nil {
			t.Fatalf("failed to update event status: %v", err)
		}

		if updatedEvent == nil {
			t.Error("expected updated event to be non-nil")
		}

		if updatedEvent.Status != "SWAPPABLE" {
			t.Errorf("expected event status to be SWAPPABLE, got %s", updatedEvent.Status)
		}
	})

	t.Run("UpdateEventStatus_Unauthorized", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		otherUser, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "unauthorized user",
			Email:    "unauthorized@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create unauthorized user: %v", err)
		}

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		createInput := CreateEventInput{
			Title:     "Unauthorized Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		createdEvent, err := eventService.CreateEvent(context.Background(), createInput)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		updateInput := UpdateEventStatusInput{
			ID:     createdEvent.ID,
			Status: "SWAPPABLE",
			UserID: otherUser.ID, // Attempt to update with unauthorized user ID
		}

		_, err = eventService.UpdateEventStatus(context.Background(), updateInput)
		if err == nil {
			t.Fatal("expected an error for unauthorized update, got nil")
		}

		if err.Error() != "user does not own this event" {
			t.Errorf("expected 'user does not own this event' error, got %v", err)
		}
	})

	t.Run("DeleteEvent", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		createInput := CreateEventInput{
			Title:     "Event to Delete",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		createdEvent, err := eventService.CreateEvent(context.Background(), createInput)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		err = eventService.DeleteEvent(context.Background(), createdEvent.ID, user.ID)
		if err != nil {
			t.Fatalf("failed to delete event: %v", err)
		}

		_, err = eventRepo.GetEventByID(context.Background(), createdEvent.ID)
		if err == nil {
			t.Fatal("expected error when getting deleted event, got nil")
		}
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})

	t.Run("DeleteEvent_Unauthorized", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		otherUser, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "unauthorized deleter",
			Email:    "unauthorized-deleter@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create unauthorized user: %v", err)
		}

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		createInput := CreateEventInput{
			Title:     "Event to Delete by Other",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "BUSY",
			UserID:    user.ID,
		}

		createdEvent, err := eventService.CreateEvent(context.Background(), createInput)
		if err != nil {
			t.Fatalf("failed to create event: %v", err)
		}

		err = eventService.DeleteEvent(context.Background(), createdEvent.ID, otherUser.ID)
		if err == nil {
			t.Fatal("expected an error for unauthorized delete, got nil")
		}

		if err.Error() != "user does not own this event" {
			t.Errorf("expected 'user does not own this event' error, got %v", err)
		}
	})

	t.Run("GetSwappableEvents", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		otherUser, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "other service user",
			Email:    "other-service-user@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create other user: %v", err)
		}

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)
		input := CreateEventInput{
			Title:     "Service Swappable Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "SWAPPABLE",
			UserID:    otherUser.ID,
		}
		_, err = eventService.CreateEvent(context.Background(), input)
		if err != nil {
			t.Fatalf("failed to create swappable event: %v", err)
		}

		swappableEvents, err := eventService.GetSwappableEvents(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("failed to get swappable events: %v", err)
		}

		if len(swappableEvents) != 1 {
			t.Errorf("expected 1 swappable event, got %d", len(swappableEvents))
		}

		if swappableEvents[0].OwnerName != otherUser.Name {
			t.Errorf("expected owner name to be %q, got %q", otherUser.Name, swappableEvents[0].OwnerName)
		}
	})

	t.Run("GetEventsByUserIDAndStatus", func(t *testing.T) {
		testQueries, user := repository.SetupTestDBWithUser(t)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)

		startTime := time.Now()
		endTime := startTime.Add(time.Hour)

		// Create a swappable event
		input := CreateEventInput{
			Title:     "Service Swappable Event",
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "SWAPPABLE",
			UserID:    user.ID,
		}
		_, err := eventService.CreateEvent(context.Background(), input)
		if err != nil {
			t.Fatalf("failed to create swappable event: %v", err)
		}

		// Create a busy event
		input.Status = "BUSY"
		_, err = eventService.CreateEvent(context.Background(), input)
		if err != nil {
			t.Fatalf("failed to create busy event: %v", err)
		}

		events, err := eventService.GetEventsByUserIDAndStatus(context.Background(), user.ID, "SWAPPABLE")
		if err != nil {
			t.Fatalf("failed to get events by user ID and status: %v", err)
		}

		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}

		if events[0].Status != "SWAPPABLE" {
			t.Errorf("expected event status to be SWAPPABLE, got %s", events[0].Status)
		}
	})

	t.Run("UpdateEvent", func(t *testing.T) {
		testQueries, user1 := repository.SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{Name: "user2", Email: "user2@example.com", Password: "password"})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventService := NewEventService(eventRepo, userRepo, swapRepo)
		swapService := NewSwapRequestService(swapRepo, eventRepo, userRepo)

		// Create events for both users
		event1, err := eventService.CreateEvent(context.Background(), CreateEventInput{Title: "Event 1", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), Status: "SWAPPABLE", UserID: user1.ID})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}
		event2, err := eventService.CreateEvent(context.Background(), CreateEventInput{Title: "Event 2", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), Status: "SWAPPABLE", UserID: user2.ID})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		// Create a swap request
		_, err = swapService.CreateSwapRequest(context.Background(), CreateSwapRequestInput{RequesterUserID: user1.ID, ResponderUserID: user2.ID, RequesterSlotID: event1.ID, ResponderSlotID: event2.ID})
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		// Update event1
		updatedTitle := "Updated Event 1"
		_, err = eventService.UpdateEvent(context.Background(), UpdateEventInput{ID: event1.ID, Title: updatedTitle, StartTime: event1.StartTime, EndTime: event1.EndTime, UserID: user1.ID})
		if err != nil {
			t.Fatalf("failed to update event: %v", err)
		}

		// Verify event1 is updated and status is BUSY
		updatedEvent1, err := eventRepo.GetEventByID(context.Background(), event1.ID)
		if err != nil {
			t.Fatalf("failed to get updated event1: %v", err)
		}
		if updatedEvent1.Title != updatedTitle {
			t.Errorf("expected updated title to be %q, got %q", updatedTitle, updatedEvent1.Title)
		}
		if updatedEvent1.Status != "BUSY" {
			t.Errorf("expected updated event status to be BUSY, got %q", updatedEvent1.Status)
		}

		// Verify event2 status is back to SWAPPABLE
		updatedEvent2, err := eventRepo.GetEventByID(context.Background(), event2.ID)
		if err != nil {
			t.Fatalf("failed to get updated event2: %v", err)
		}
		if updatedEvent2.Status != "SWAPPABLE" {
			t.Errorf("expected other event status to be SWAPPABLE, got %q", updatedEvent2.Status)
		}

		// Verify swap request is deleted
		swapRequests, err := swapRepo.GetSwapRequestsByEventID(context.Background(), event1.ID)
		if err != nil {
			t.Fatalf("failed to get swap requests by event ID: %v", err)
		}
		if len(swapRequests) != 0 {
			t.Errorf("expected swap requests to be deleted, but found %d", len(swapRequests))
		}
	})
}
