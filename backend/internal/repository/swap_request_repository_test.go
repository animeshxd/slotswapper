package repository

import (
	"context"
	"testing"
	"time"

	"slotswapper/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

func TestSwapRequestRepository(t *testing.T) {
	swapRepo := NewSwapRequestRepository(nil) // Will be initialized in sub-tests

	t.Run("CreateSwapRequest", func(t *testing.T) {
		testQueries, user1 := SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2",
			Email:    "user2@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo = NewSwapRequestRepository(testQueries)

		arg := db.CreateSwapRequestParams{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
			Status:          "PENDING",
		}

		swapRequest, err := swapRepo.CreateSwapRequest(context.Background(), arg)
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		if swapRequest.ID == 0 {
			t.Error("expected swap request ID to be non-zero")
		}
		if swapRequest.Status != arg.Status {
			t.Errorf("expected status %q, got %q", arg.Status, swapRequest.Status)
		}
	})

	t.Run("GetSwapRequestByID", func(t *testing.T) {
		testQueries, user1 := SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_getbyid",
			Email:    "user2_getbyid@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event GetByID",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event GetByID",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo = NewSwapRequestRepository(testQueries)

		createArg := db.CreateSwapRequestParams{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
			Status:          "PENDING",
		}

		createdSwapRequest, err := swapRepo.CreateSwapRequest(context.Background(), createArg)
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		retrievedSwapRequest, err := swapRepo.GetSwapRequestByID(context.Background(), createdSwapRequest.ID)
		if err != nil {
			t.Fatalf("failed to get swap request by ID: %v", err)
		}

		if retrievedSwapRequest.ID != createdSwapRequest.ID {
			t.Errorf("expected swap request ID %d, got %d", createdSwapRequest.ID, retrievedSwapRequest.ID)
		}
	})

	t.Run("GetIncomingSwapRequests", func(t *testing.T) {
		testQueries, user1 := SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_incoming",
			Email:    "user2_incoming@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event Incoming",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event Incoming",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo = NewSwapRequestRepository(testQueries)

		createArg := db.CreateSwapRequestParams{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
			Status:          "PENDING",
		}
		_, err = swapRepo.CreateSwapRequest(context.Background(), createArg)
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		incomingRequests, err := swapRepo.GetIncomingSwapRequests(context.Background(), user2.ID)
		if err != nil {
			t.Fatalf("failed to get incoming swap requests: %v", err)
		}

		if len(incomingRequests) != 1 {
			t.Errorf("expected 1 incoming swap request, got %d", len(incomingRequests))
		}
		if incomingRequests[0].RequesterUserID != user1.ID {
			t.Errorf("expected requester user ID %d, got %d", user1.ID, incomingRequests[0].RequesterUserID)
		}
	})

	t.Run("GetOutgoingSwapRequests", func(t *testing.T) {
		testQueries, user1 := SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_outgoing",
			Email:    "user2_outgoing@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event Outgoing",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event Outgoing",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo = NewSwapRequestRepository(testQueries)

		createArg := db.CreateSwapRequestParams{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
			Status:          "PENDING",
		}
		_, err = swapRepo.CreateSwapRequest(context.Background(), createArg)
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		outgoingRequests, err := swapRepo.GetOutgoingSwapRequests(context.Background(), user1.ID)
		if err != nil {
			t.Fatalf("failed to get outgoing swap requests: %v", err)
		}

		if len(outgoingRequests) != 1 {
			t.Errorf("expected 1 outgoing swap request, got %d", len(outgoingRequests))
		}
		if outgoingRequests[0].ResponderUserID != user2.ID {
			t.Errorf("expected responder user ID %d, got %d", user2.ID, outgoingRequests[0].ResponderUserID)
		}
	})

	t.Run("UpdateSwapRequestStatus", func(t *testing.T) {
		testQueries, user1 := SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_update",
			Email:    "user2_update@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event Update",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event Update",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo = NewSwapRequestRepository(testQueries)

		createArg := db.CreateSwapRequestParams{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
			Status:          "PENDING",
		}

		createdSwapRequest, err := swapRepo.CreateSwapRequest(context.Background(), createArg)
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		updateArg := db.UpdateSwapRequestStatusParams{
			ID:     createdSwapRequest.ID,
			Status: "ACCEPTED",
		}

		updatedSwapRequest, err := swapRepo.UpdateSwapRequestStatus(context.Background(), updateArg)
		if err != nil {
			t.Fatalf("failed to update swap request status: %v", err)
		}

		if updatedSwapRequest.Status != "ACCEPTED" {
			t.Errorf("expected status ACCEPTED, got %q", updatedSwapRequest.Status)
		}
	})
}
