package software

import (
	"context"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("software not found")

// Service is the business logic layer for the software catalog.
type Service struct{ repo SoftwareRepo }

func NewService(repo SoftwareRepo) *Service { return &Service{repo: repo} }

// ─── Core software ────────────────────────────────────────────────────────────

func (s *Service) List(ctx context.Context, p ListParams) ([]Software, int64, error) {
	return s.repo.List(ctx, p)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Software, error) {
	sw, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return sw, nil
}

func (s *Service) Create(ctx context.Context, input CreateSoftwareInput) (*Software, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateSoftwareInput) (*Software, error) {
	sw, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return sw, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetInhouseDetails(ctx context.Context, id string) (*InhouseDetails, error) {
	return s.repo.GetInhouseDetails(ctx, id)
}

func (s *Service) UpsertInhouseDetails(ctx context.Context, id string, d InhouseDetails) error {
	return s.repo.UpsertInhouseDetails(ctx, id, d)
}

func (s *Service) GetVendorDetails(ctx context.Context, id string) (*VendorDetails, error) {
	return s.repo.GetVendorDetails(ctx, id)
}

func (s *Service) UpsertVendorDetails(ctx context.Context, id string, d VendorDetails) error {
	return s.repo.UpsertVendorDetails(ctx, id, d)
}

// ─── Sub-resources ────────────────────────────────────────────────────────────

func (s *Service) ListResponsibilities(ctx context.Context, id string) ([]Responsibility, error) {
	return s.repo.ListResponsibilities(ctx, id)
}
func (s *Service) AddResponsibility(ctx context.Context, id string, input AddResponsibilityInput) (*Responsibility, error) {
	return s.repo.AddResponsibility(ctx, id, input)
}
func (s *Service) DeleteResponsibility(ctx context.Context, softwareID, respID string) error {
	return s.repo.DeleteResponsibility(ctx, softwareID, respID)
}

func (s *Service) ListRepos(ctx context.Context, id string) ([]Repository, error) {
	return s.repo.ListRepos(ctx, id)
}
func (s *Service) AddRepo(ctx context.Context, id string, input AddRepoInput) (*Repository, error) {
	return s.repo.AddRepo(ctx, id, input)
}
func (s *Service) DeleteRepo(ctx context.Context, softwareID, repoID string) error {
	return s.repo.DeleteRepo(ctx, softwareID, repoID)
}

func (s *Service) ListDeployments(ctx context.Context, id string) ([]Deployment, error) {
	return s.repo.ListDeployments(ctx, id)
}
func (s *Service) AddDeployment(ctx context.Context, id string, input AddDeploymentInput) (*Deployment, error) {
	return s.repo.AddDeployment(ctx, id, input)
}
func (s *Service) UpdateDeploymentHealth(ctx context.Context, softwareID, depID, health string) error {
	return s.repo.UpdateDeploymentHealth(ctx, softwareID, depID, health)
}
func (s *Service) DeleteDeployment(ctx context.Context, softwareID, depID string) error {
	return s.repo.DeleteDeployment(ctx, softwareID, depID)
}

func (s *Service) ListTechStack(ctx context.Context, id string) ([]TechStack, error) {
	return s.repo.ListTechStack(ctx, id)
}
func (s *Service) AddTechStack(ctx context.Context, id string, input AddTechStackInput) (*TechStack, error) {
	return s.repo.AddTechStack(ctx, id, input)
}
func (s *Service) DeleteTechStack(ctx context.Context, softwareID, techID string) error {
	return s.repo.DeleteTechStack(ctx, softwareID, techID)
}

func (s *Service) ListDatabaseLinks(ctx context.Context, id string) ([]DatabaseLink, error) {
	return s.repo.ListDatabaseLinks(ctx, id)
}
func (s *Service) AddDatabaseLink(ctx context.Context, id string, input AddDatabaseLinkInput) (*DatabaseLink, error) {
	return s.repo.AddDatabaseLink(ctx, id, input)
}
func (s *Service) DeleteDatabaseLink(ctx context.Context, softwareID, linkID string) error {
	return s.repo.DeleteDatabaseLink(ctx, softwareID, linkID)
}
