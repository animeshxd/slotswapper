package repository

import (
	"context"

	"slotswapper/internal/db"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, arg db.CreateEventParams) (db.Event, error)
	GetEventByID(ctx context.Context, id int64) (db.Event, error)
	GetEventsByUserID(ctx context.Context, userID int64) ([]db.Event, error)
	GetEventsByUserIDAndStatus(ctx context.Context, params db.GetEventsByUserIDAndStatusParams) ([]db.Event, error)
	UpdateEventStatus(ctx context.Context, arg db.UpdateEventStatusParams) (db.Event, error)
	UpdateEventUserID(ctx context.Context, arg db.UpdateEventUserIDParams) (db.Event, error)
	DeleteEvent(ctx context.Context, id int64) error
	GetSwappableEvents(ctx context.Context, userID int64) ([]db.GetSwappableEventsRow, error)
}

func (r *eventRepository) DeleteEvent(ctx context.Context, id int64) error {
	return r.queries.DeleteEvent(ctx, id)
}

type eventRepository struct {
	queries *db.Queries
}

func NewEventRepository(queries *db.Queries) EventRepository {
	return &eventRepository{queries: queries}
}

func (r *eventRepository) CreateEvent(ctx context.Context, arg db.CreateEventParams) (db.Event, error) {
	return r.queries.CreateEvent(ctx, arg)
}

func (r *eventRepository) GetEventByID(ctx context.Context, id int64) (db.Event, error) {
	return r.queries.GetEventByID(ctx, id)
}

func (r *eventRepository) GetEventsByUserID(ctx context.Context, userID int64) ([]db.Event, error) {
	return r.queries.GetEventsByUserID(ctx, userID)
}

func (r *eventRepository) GetEventsByUserIDAndStatus(ctx context.Context, params db.GetEventsByUserIDAndStatusParams) ([]db.Event, error) {
	return r.queries.GetEventsByUserIDAndStatus(ctx, params)
}

func (r *eventRepository) UpdateEventStatus(ctx context.Context, arg db.UpdateEventStatusParams) (db.Event, error) {
	return r.queries.UpdateEventStatus(ctx, arg)
}

func (r *eventRepository) UpdateEventUserID(ctx context.Context, arg db.UpdateEventUserIDParams) (db.Event, error) {
	return r.queries.UpdateEventUserID(ctx, arg)
}

func (r *eventRepository) GetSwappableEvents(ctx context.Context, userID int64) ([]db.GetSwappableEventsRow, error) {
	return r.queries.GetSwappableEvents(ctx, userID)
}
