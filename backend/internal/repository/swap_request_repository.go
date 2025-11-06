package repository

import (
	"context"

	"slotswapper/internal/db"
)

type SwapRequestRepository interface {
	CreateSwapRequest(ctx context.Context, arg db.CreateSwapRequestParams) (db.SwapRequest, error)
	GetSwapRequestByID(ctx context.Context, id int64) (db.SwapRequest, error)
	GetIncomingSwapRequests(ctx context.Context, userID int64) ([]db.GetIncomingSwapRequestsRow, error)
	GetOutgoingSwapRequests(ctx context.Context, requesterUserID int64) ([]db.GetOutgoingSwapRequestsRow, error)
	UpdateSwapRequestStatus(ctx context.Context, arg db.UpdateSwapRequestStatusParams) (db.SwapRequest, error)
	DeleteSwapRequest(ctx context.Context, id int64) error
	GetSwapRequestsByEventID(ctx context.Context, eventID int64) ([]db.SwapRequest, error)
}

type swapRequestRepository struct {
	queries *db.Queries
}

func NewSwapRequestRepository(queries *db.Queries) SwapRequestRepository {
	return &swapRequestRepository{queries: queries}
}

func (r *swapRequestRepository) CreateSwapRequest(ctx context.Context, arg db.CreateSwapRequestParams) (db.SwapRequest, error) {
	return r.queries.CreateSwapRequest(ctx, arg)
}

func (r *swapRequestRepository) GetSwapRequestByID(ctx context.Context, id int64) (db.SwapRequest, error) {
	return r.queries.GetSwapRequestByID(ctx, id)
}

func (r *swapRequestRepository) GetIncomingSwapRequests(ctx context.Context, userID int64) ([]db.GetIncomingSwapRequestsRow, error) {
	return r.queries.GetIncomingSwapRequests(ctx, userID)
}

func (r *swapRequestRepository) GetOutgoingSwapRequests(ctx context.Context, requesterUserID int64) ([]db.GetOutgoingSwapRequestsRow, error) {
	return r.queries.GetOutgoingSwapRequests(ctx, requesterUserID)
}

func (r *swapRequestRepository) UpdateSwapRequestStatus(ctx context.Context, arg db.UpdateSwapRequestStatusParams) (db.SwapRequest, error) {
	return r.queries.UpdateSwapRequestStatus(ctx, arg)
}

func (r *swapRequestRepository) DeleteSwapRequest(ctx context.Context, id int64) error {
	return r.queries.DeleteSwapRequest(ctx, id)
}

func (r *swapRequestRepository) GetSwapRequestsByEventID(ctx context.Context, eventID int64) ([]db.SwapRequest, error) {
	return r.queries.GetSwapRequestsByEventID(ctx, db.GetSwapRequestsByEventIDParams{RequesterSlotID: eventID, ResponderSlotID: eventID})
}
