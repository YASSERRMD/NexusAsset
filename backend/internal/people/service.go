package people

import (
	"context"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("record not found")

// PersonService handles persons business logic.
type PersonService struct{ repo PersonRepository }

func NewPersonService(repo PersonRepository) *PersonService { return &PersonService{repo: repo} }

func (s *PersonService) List(ctx context.Context, p ListParams) ([]Person, int64, error) {
	return s.repo.List(ctx, p)
}
func (s *PersonService) GetByID(ctx context.Context, id string) (*Person, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return p, nil
}
func (s *PersonService) Create(ctx context.Context, input CreatePersonInput) (*Person, error) {
	return s.repo.Create(ctx, input)
}
func (s *PersonService) Update(ctx context.Context, id string, input UpdatePersonInput) (*Person, error) {
	p, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return p, nil
}
func (s *PersonService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// TeamService handles teams business logic.
type TeamService struct{ repo TeamRepository }

func NewTeamService(repo TeamRepository) *TeamService { return &TeamService{repo: repo} }

func (s *TeamService) List(ctx context.Context, p ListParams) ([]Team, int64, error) {
	return s.repo.List(ctx, p)
}
func (s *TeamService) GetByID(ctx context.Context, id string) (*Team, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return t, nil
}
func (s *TeamService) Create(ctx context.Context, input CreateTeamInput) (*Team, error) {
	return s.repo.Create(ctx, input)
}
func (s *TeamService) Update(ctx context.Context, id string, input UpdateTeamInput) (*Team, error) {
	t, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return t, nil
}
func (s *TeamService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
