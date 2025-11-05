package services

import (
	"context"
	"errors"

	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/validation"
)

type CreateSwapRequestInput struct {
	RequesterUserID int64 `json:"requester_user_id" validate:"required"`
	ResponderUserID int64 `json:"responder_user_id" validate:"required"`
	RequesterSlotID int64 `json:"requester_slot_id" validate:"required"`
	ResponderSlotID int64 `json:"responder_slot_id" validate:"required"`
}

type UpdateSwapRequestStatusInput struct {
	ID     int64  `json:"id" validate:"required"`
	Status string `json:"status" validate:"required,oneof=PENDING ACCEPTED REJECTED"`
	UserID int64  `json:"user_id" validate:"required"` // User performing the update
}

type SwapRequestService interface {
	CreateSwapRequest(ctx context.Context, input CreateSwapRequestInput) (*db.SwapRequest, error)
	GetSwapRequestByID(ctx context.Context, id int64) (*db.SwapRequest, error)
	GetIncomingSwapRequests(ctx context.Context, responderUserID int64) ([]db.SwapRequest, error)
	GetOutgoingSwapRequests(ctx context.Context, requesterUserID int64) ([]db.SwapRequest, error)
	UpdateSwapRequestStatus(ctx context.Context, input UpdateSwapRequestStatusInput) (*db.SwapRequest, error)
}

type swapRequestService struct {
	swapRepo  repository.SwapRequestRepository
	eventRepo repository.EventRepository
	userRepo  repository.UserRepository
}

func NewSwapRequestService(swapRepo repository.SwapRequestRepository, eventRepo repository.EventRepository, userRepo repository.UserRepository) SwapRequestService {
	return &swapRequestService{swapRepo: swapRepo, eventRepo: eventRepo, userRepo: userRepo}
}

func (s *swapRequestService) CreateSwapRequest(ctx context.Context, input CreateSwapRequestInput) (*db.SwapRequest, error) {
	if err := validation.Validate.Struct(input); err != nil {
		return nil, err
	}

	if input.RequesterUserID == input.ResponderUserID {
		return nil, errors.New("cannot swap with yourself")
	}

	requesterEvent, err := s.eventRepo.GetEventByID(ctx, input.RequesterSlotID)
	if err != nil {
		return nil, errors.New("requester slot not found")
	}
	if requesterEvent.Status != "SWAPPABLE" {
		return nil, errors.New("requester slot is not swappable")
	}
	if requesterEvent.UserID != input.RequesterUserID {
		return nil, errors.New("requester does not own the requester slot")
	}

	responderEvent, err := s.eventRepo.GetEventByID(ctx, input.ResponderSlotID)
	if err != nil {
		return nil, errors.New("responder slot not found")
	}
	if responderEvent.Status != "SWAPPABLE" {
		return nil, errors.New("responder slot is not swappable")
	}
	if responderEvent.UserID != input.ResponderUserID {
		return nil, errors.New("responder does not own the responder slot")
	}

	_, err = s.eventRepo.UpdateEventStatus(ctx, db.UpdateEventStatusParams{
		ID:     requesterEvent.ID,
		Status: "SWAP_PENDING",
	})
	if err != nil {
		return nil, err
	}

	_, err = s.eventRepo.UpdateEventStatus(ctx, db.UpdateEventStatusParams{
		ID:     responderEvent.ID,
		Status: "SWAP_PENDING",
	})
	if err != nil {
		return nil, err
	}

	arg := db.CreateSwapRequestParams{
		RequesterUserID: input.RequesterUserID,
		ResponderUserID: input.ResponderUserID,
		RequesterSlotID: input.RequesterSlotID,
		ResponderSlotID: input.ResponderSlotID,
		Status:          "PENDING",
	}

	swapRequest, err := s.swapRepo.CreateSwapRequest(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &swapRequest, nil
}

func (s *swapRequestService) GetSwapRequestByID(ctx context.Context, id int64) (*db.SwapRequest, error) {
	swapRequest, err := s.swapRepo.GetSwapRequestByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &swapRequest, nil
}

func (s *swapRequestService) GetIncomingSwapRequests(ctx context.Context, responderUserID int64) ([]db.SwapRequest, error) {
	return s.swapRepo.GetIncomingSwapRequests(ctx, responderUserID)
}

func (s *swapRequestService) GetOutgoingSwapRequests(ctx context.Context, requesterUserID int64) ([]db.SwapRequest, error) {
	return s.swapRepo.GetOutgoingSwapRequests(ctx, requesterUserID)
}

func (s *swapRequestService) UpdateSwapRequestStatus(ctx context.Context, input UpdateSwapRequestStatusInput) (*db.SwapRequest, error) {
	if err := validation.Validate.Struct(input); err != nil {
		return nil, err
	}

	swapRequest, err := s.swapRepo.GetSwapRequestByID(ctx, input.ID)
	if err != nil {
		return nil, errors.New("swap request not found")
	}

	if swapRequest.Status != "PENDING" {
		return nil, errors.New("swap request is not in PENDING status")
	}

	if swapRequest.ResponderUserID != input.UserID {
		return nil, errors.New("user is not authorized to update this swap request")
	}

	switch input.Status {
	case "REJECTED":
		_, err = s.eventRepo.UpdateEventStatus(ctx, db.UpdateEventStatusParams{
			ID:     swapRequest.RequesterSlotID,
			Status: "SWAPPABLE",
		})
		if err != nil {
			return nil, err
		}
		_, err = s.eventRepo.UpdateEventStatus(ctx, db.UpdateEventStatusParams{
			ID:     swapRequest.ResponderSlotID,
			Status: "SWAPPABLE",
		})
		if err != nil {
			return nil, err
		}
	case "ACCEPTED":

		requesterEvent, err := s.eventRepo.GetEventByID(ctx, swapRequest.RequesterSlotID)
		if err != nil {
			return nil, err
		}
		responderEvent, err := s.eventRepo.GetEventByID(ctx, swapRequest.ResponderSlotID)
		if err != nil {
			return nil, err
		}

		_, err = s.eventRepo.UpdateEventUserID(ctx, db.UpdateEventUserIDParams{
			ID:     requesterEvent.ID,
			UserID: responderEvent.UserID,
		})
		if err != nil {
			return nil, err
		}
		_, err = s.eventRepo.UpdateEventUserID(ctx, db.UpdateEventUserIDParams{
			ID:     responderEvent.ID,
			UserID: requesterEvent.UserID,
		})
		if err != nil {
			return nil, err
		}

		_, err = s.eventRepo.UpdateEventStatus(ctx, db.UpdateEventStatusParams{
			ID:     requesterEvent.ID,
			Status: "BUSY",
		})
		if err != nil {
			return nil, err
		}
		_, err = s.eventRepo.UpdateEventStatus(ctx, db.UpdateEventStatusParams{
			ID:     responderEvent.ID,
			Status: "BUSY",
		})
		if err != nil {
			return nil, err
		}
	}
	arg := db.UpdateSwapRequestStatusParams{
		ID:     input.ID,
		Status: input.Status,
	}

	updatedSwapRequest, err := s.swapRepo.UpdateSwapRequestStatus(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &updatedSwapRequest, nil
}
