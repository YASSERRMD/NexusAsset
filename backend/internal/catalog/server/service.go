package server

import (
	"context"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("server not found")

// Service is the business logic layer for the server catalog.
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, p ListParams) ([]Server, int64, error) {
	return s.repo.List(ctx, p)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Server, error) {
	srv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return srv, nil
}

func (s *Service) Create(ctx context.Context, input CreateServerInput) (*Server, error) {
	if input.Status == "" {
		input.Status = "active"
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateServerInput) (*Server, error) {
	srv, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return srv, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
