package services

import (
	"context"
	"testing"
	"time"

	"slotswapper/internal/db"
	"slotswapper/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func TestSwapRequestService(t *testing.T) {
	t.Run("CreateSwapRequest", func(t *testing.T) {
		testQueries, user1 := repository.SetupTestDBWithUser(t)
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

		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapService := NewSwapRequestService(swapRepo, eventRepo, userRepo)

		input := CreateSwapRequestInput{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
		}

		swapRequest, err := swapService.CreateSwapRequest(context.Background(), input)
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		if swapRequest == nil {
			t.Error("expected swap request to be non-nil")
		}

		if swapRequest.ID == 0 {
			t.Error("expected swap request ID to be non-zero")
		}
		if swapRequest.Status != "PENDING" {
			t.Errorf("expected status PENDING, got %q", swapRequest.Status)
		}

		updatedEvent1, err := eventRepo.GetEventByID(context.Background(), event1.ID)
		if err != nil {
			t.Fatalf("failed to get updated event1: %v", err)
		}
		if updatedEvent1.Status != "SWAP_PENDING" {
			t.Errorf("expected event1 status to be SWAP_PENDING, got %q", updatedEvent1.Status)
		}

		updatedEvent2, err := eventRepo.GetEventByID(context.Background(), event2.ID)
		if err != nil {
			t.Fatalf("failed to get updated event2: %v", err)
		}
		if updatedEvent2.Status != "SWAP_PENDING" {
			t.Errorf("expected event2 status to be SWAP_PENDING, got %q", updatedEvent2.Status)
		}
	})

	t.Run("CreateSwapRequest_ValidationErrors", func(t *testing.T) {
		testQueries, user1 := repository.SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_val",
			Email:    "user2_val@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event Val",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event Val",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapService := NewSwapRequestService(swapRepo, eventRepo, userRepo)

		testCases := []struct {
			name          string
			input         CreateSwapRequestInput
			expectedError string
		}{
			{
				name: "missing requester user ID",
				input: CreateSwapRequestInput{
					ResponderUserID: user2.ID,
					RequesterSlotID: event1.ID,
					ResponderSlotID: event2.ID,
				},
				expectedError: "Key: 'CreateSwapRequestInput.RequesterUserID' Error:Field validation for 'RequesterUserID' failed on the 'required' tag",
			},
			{
				name: "missing responder user ID",
				input: CreateSwapRequestInput{
					RequesterUserID: user1.ID,
					RequesterSlotID: event1.ID,
					ResponderSlotID: event2.ID,
				},
				expectedError: "Key: 'CreateSwapRequestInput.ResponderUserID' Error:Field validation for 'ResponderUserID' failed on the 'required' tag",
			},
			{
				name: "missing requester slot ID",
				input: CreateSwapRequestInput{
					RequesterUserID: user1.ID,
					ResponderUserID: user2.ID,
					ResponderSlotID: event2.ID,
				},
				expectedError: "Key: 'CreateSwapRequestInput.RequesterSlotID' Error:Field validation for 'RequesterSlotID' failed on the 'required' tag",
			},
			{
				name: "missing responder slot ID",
				input: CreateSwapRequestInput{
					RequesterUserID: user1.ID,
					ResponderUserID: user2.ID,
					RequesterSlotID: event1.ID,
				},
				expectedError: "Key: 'CreateSwapRequestInput.ResponderSlotID' Error:Field validation for 'ResponderSlotID' failed on the 'required' tag",
			},
			{
				name: "swap with self",
				input: CreateSwapRequestInput{
					RequesterUserID: user1.ID,
					ResponderUserID: user1.ID,
					RequesterSlotID: event1.ID,
					ResponderSlotID: event2.ID,
				},
				expectedError: "cannot swap with yourself",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := swapService.CreateSwapRequest(context.Background(), tc.input)
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if err.Error() != tc.expectedError {
					t.Errorf("expected error %q, got %q", tc.expectedError, err.Error())
				}
			})
		}
	})

	t.Run("UpdateSwapRequestStatus_Accepted", func(t *testing.T) {
		testQueries, user1 := repository.SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_accept",
			Email:    "user2_accept@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event Accept",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event Accept",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapService := NewSwapRequestService(swapRepo, eventRepo, userRepo)

		createInput := CreateSwapRequestInput{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
		}

		createdSwapRequest, err := swapService.CreateSwapRequest(context.Background(), createInput)
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		updateInput := UpdateSwapRequestStatusInput{
			ID:     createdSwapRequest.ID,
			Status: "ACCEPTED",
			UserID: user2.ID, // Responder is accepting
		}

		updatedSwapRequest, err := swapService.UpdateSwapRequestStatus(context.Background(), updateInput)
		if err != nil {
			t.Fatalf("failed to update swap request status: %v", err)
		}

		if updatedSwapRequest == nil {
			t.Error("expected updated swap request to be non-nil")
		}

		if updatedSwapRequest.Status != "ACCEPTED" {
			t.Errorf("expected status ACCEPTED, got %q", updatedSwapRequest.Status)
		}

		finalEvent1, err := eventRepo.GetEventByID(context.Background(), event1.ID)
		if err != nil {
			t.Fatalf("failed to get final event1: %v", err)
		}
		if finalEvent1.UserID != user2.ID {
			t.Errorf("expected final event1 user ID to be %d, got %d", user2.ID, finalEvent1.UserID)
		}
		if finalEvent1.Status != "BUSY" {
			t.Errorf("expected final event1 status to be BUSY, got %q", finalEvent1.Status)
		}

		finalEvent2, err := eventRepo.GetEventByID(context.Background(), event2.ID)
		if err != nil {
			t.Fatalf("failed to get final event2: %v", err)
		}
		if finalEvent2.UserID != user1.ID {
			t.Errorf("expected final event2 user ID to be %d, got %d", user1.ID, finalEvent2.UserID)
		}
		if finalEvent2.Status != "BUSY" {
			t.Errorf("expected final event2 status to be BUSY, got %q", finalEvent2.Status)
		}
	})

	t.Run("UpdateSwapRequestStatus_Rejected", func(t *testing.T) {
		testQueries, user1 := repository.SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_reject",
			Email:    "user2_reject@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event Reject",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAP_PENDING", // Already in SWAP_PENDING state
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event Reject",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAP_PENDING", // Already in SWAP_PENDING state
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		createdSwapRequest, err := testQueries.CreateSwapRequest(context.Background(), db.CreateSwapRequestParams{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
			Status:          "PENDING",
		})
		if err != nil {
			t.Fatalf("failed to create swap request for rejection test: %v", err)
		}

		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapService := NewSwapRequestService(swapRepo, eventRepo, userRepo)

		updateInput := UpdateSwapRequestStatusInput{
			ID:     createdSwapRequest.ID,
			Status: "REJECTED",
			UserID: user2.ID, // Responder is rejecting
		}

		updatedSwapRequest, err := swapService.UpdateSwapRequestStatus(context.Background(), updateInput)
		if err != nil {
			t.Fatalf("failed to update swap request status: %v", err)
		}

		if updatedSwapRequest == nil {
			t.Error("expected updated swap request to be non-nil")
		}

		if updatedSwapRequest.Status != "REJECTED" {
			t.Errorf("expected status REJECTED, got %q", updatedSwapRequest.Status)
		}

		finalEvent1, err := eventRepo.GetEventByID(context.Background(), event1.ID)
		if err != nil {
			t.Fatalf("failed to get final event1: %v", err)
		}
		if finalEvent1.Status != "SWAPPABLE" {
			t.Errorf("expected final event1 status to be SWAPPABLE, got %q", finalEvent1.Status)
		}

		finalEvent2, err := eventRepo.GetEventByID(context.Background(), event2.ID)
		if err != nil {
			t.Fatalf("failed to get final event2: %v", err)
		}
		if finalEvent2.Status != "SWAPPABLE" {
			t.Errorf("expected final event2 status to be SWAPPABLE, got %q", finalEvent2.Status)
		}
	})

	t.Run("UpdateSwapRequestStatus_NotPending", func(t *testing.T) {
		testQueries, user1 := repository.SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_notpending",
			Email:    "user2_notpending@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Event NotPending",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Event NotPending",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		createdSwapRequest, err := testQueries.CreateSwapRequest(context.Background(), db.CreateSwapRequestParams{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
			Status:          "ACCEPTED",
		})
		if err != nil {
			t.Fatalf("failed to create swap request for not pending test: %v", err)
		}

		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapService := NewSwapRequestService(swapRepo, eventRepo, userRepo)

		updateInput := UpdateSwapRequestStatusInput{
			ID:     createdSwapRequest.ID,
			Status: "REJECTED",
			UserID: user2.ID, // Responder is rejecting
		}

		_, err = swapService.UpdateSwapRequestStatus(context.Background(), updateInput)
		if err == nil {
			t.Fatal("expected an error for updating non-pending swap request, got nil")
		}

		if err.Error() != "swap request is not in PENDING status" {
			t.Errorf("expected 'swap request is not in PENDING status' error, got %q", err)
		}
	})

	t.Run("GetIncomingSwapRequests", func(t *testing.T) {
		testQueries, user1 := repository.SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_incoming",
			Email:    "user2_incoming@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Incoming Event",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Incoming Event",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapService := NewSwapRequestService(swapRepo, eventRepo, userRepo)

		_, err = swapService.CreateSwapRequest(context.Background(), CreateSwapRequestInput{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		incomingRequests, err := swapService.GetIncomingSwapRequests(context.Background(), user2.ID)
		if err != nil {
			t.Fatalf("failed to get incoming swap requests: %v", err)
		}

		if len(incomingRequests) != 1 {
			t.Errorf("expected 1 incoming request, got %d", len(incomingRequests))
		}

		if incomingRequests[0].RequesterName != user1.Name {
			t.Errorf("expected requester name %q, got %q", user1.Name, incomingRequests[0].RequesterName)
		}
		if incomingRequests[0].RequesterEventTitle != event1.Title {
			t.Errorf("expected requester event title %q, got %q", event1.Title, incomingRequests[0].RequesterEventTitle)
		}
	})

	t.Run("GetOutgoingSwapRequests", func(t *testing.T) {
		testQueries, user1 := repository.SetupTestDBWithUser(t)
		user2, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
			Name:     "user2_outgoing",
			Email:    "user2_outgoing@example.com",
			Password: "password",
		})
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		event1, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User1 Outgoing Event",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user1.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event1: %v", err)
		}

		event2, err := testQueries.CreateEvent(context.Background(), db.CreateEventParams{
			Title:     "User2 Outgoing Event",
			StartTime: time.Now().Add(2 * time.Hour),
			EndTime:   time.Now().Add(3 * time.Hour),
			Status:    "SWAPPABLE",
			UserID:    user2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create event2: %v", err)
		}

		swapRepo := repository.NewSwapRequestRepository(testQueries)
		eventRepo := repository.NewEventRepository(testQueries)
		userRepo := repository.NewUserRepository(testQueries)
		swapService := NewSwapRequestService(swapRepo, eventRepo, userRepo)

		_, err = swapService.CreateSwapRequest(context.Background(), CreateSwapRequestInput{
			RequesterUserID: user1.ID,
			ResponderUserID: user2.ID,
			RequesterSlotID: event1.ID,
			ResponderSlotID: event2.ID,
		})
		if err != nil {
			t.Fatalf("failed to create swap request: %v", err)
		}

		outgoingRequests, err := swapService.GetOutgoingSwapRequests(context.Background(), user1.ID)
		if err != nil {
			t.Fatalf("failed to get outgoing swap requests: %v", err)
		}

		if len(outgoingRequests) != 1 {
			t.Errorf("expected 1 outgoing request, got %d", len(outgoingRequests))
		}

		if outgoingRequests[0].ResponderName != user2.Name {
			t.Errorf("expected responder name %q, got %q", user2.Name, outgoingRequests[0].ResponderName)
		}
		if outgoingRequests[0].RequesterEventTitle != event1.Title {
			t.Errorf("expected requester event title %q, got %q", event1.Title, outgoingRequests[0].RequesterEventTitle)
		}
	})

}
