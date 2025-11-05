package services

import (
	"context"
	"errors"
	"time"

	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/validation"
)

type CreateEventInput struct {
	Title     string    `json:"title" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	Status    string    `json:"status" validate:"required,oneof=BUSY SWAPPABLE SWAP_PENDING"`
	UserID    int64     `json:"user_id" validate:"required"`
}

type UpdateEventStatusInput struct {
	ID     int64  `json:"id" validate:"required"`
	Status string `json:"status" validate:"required,oneof=BUSY SWAPPABLE SWAP_PENDING"`
	UserID int64  `json:"user_id" validate:"required"` // User performing the update
}

type EventService interface {
	CreateEvent(ctx context.Context, input CreateEventInput) (*db.Event, error)
	GetEventByID(ctx context.Context, id int64) (*db.Event, error)
	GetEventsByUserID(ctx context.Context, userID int64) ([]db.Event, error)
	UpdateEventStatus(ctx context.Context, input UpdateEventStatusInput) (*db.Event, error)
	DeleteEvent(ctx context.Context, eventID, userID int64) error
	GetSwappableEvents(ctx context.Context, userID int64) ([]db.Event, error)
}

func (s *eventService) DeleteEvent(ctx context.Context, eventID, userID int64) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return errors.New("event not found")
	}

	if event.UserID != userID {
		return errors.New("user does not own this event")
	}

	return s.eventRepo.DeleteEvent(ctx, eventID)
}

type eventService struct {
	eventRepo repository.EventRepository
	userRepo  repository.UserRepository
}

func NewEventService(eventRepo repository.EventRepository, userRepo repository.UserRepository) EventService {
	return &eventService{eventRepo: eventRepo, userRepo: userRepo}
}

func (s *eventService) CreateEvent(ctx context.Context, input CreateEventInput) (*db.Event, error) {
	if err := validation.Validate.Struct(input); err != nil {
		return nil, err
	}

	arg := db.CreateEventParams{
		Title:     input.Title,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		Status:    input.Status,
		UserID:    input.UserID,
	}

	event, err := s.eventRepo.CreateEvent(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *eventService) GetEventByID(ctx context.Context, id int64) (*db.Event, error) {
	event, err := s.eventRepo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *eventService) GetEventsByUserID(ctx context.Context, userID int64) ([]db.Event, error) {
	return s.eventRepo.GetEventsByUserID(ctx, userID)
}

func (s *eventService) UpdateEventStatus(ctx context.Context, input UpdateEventStatusInput) (*db.Event, error) {
	if err := validation.Validate.Struct(input); err != nil {
		return nil, err
	}

	event, err := s.eventRepo.GetEventByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if event.UserID != input.UserID {
		return nil, errors.New("user does not own this event")
	}

	arg := db.UpdateEventStatusParams{
		ID:     input.ID,
		Status: input.Status,
	}

	updatedEvent, err := s.eventRepo.UpdateEventStatus(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &updatedEvent, nil
}

func (s *eventService) GetSwappableEvents(ctx context.Context, userID int64) ([]db.Event, error) {
	return s.eventRepo.GetSwappableEvents(ctx, userID)
}
