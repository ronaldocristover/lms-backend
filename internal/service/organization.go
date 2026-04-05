package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/internal/repository"
	"github.com/ronaldocristover/lms-backend/pkg/apierror"
)

var (
	ErrOrganizationNotFound      = errors.New("organization not found")
	ErrOrganizationExists        = errors.New("organization with this name already exists")
	ErrUserAlreadyInOrg          = errors.New("user already in organization")
	ErrUserNotInOrg              = errors.New("user not in organization")
	ErrCannotRemoveOwner         = errors.New("cannot remove organization owner")
	ErrInvalidOrganizationID     = errors.New("invalid organization ID")
	ErrInvalidUserID             = errors.New("invalid user ID")
	ErrOwnerNotFound             = errors.New("owner not found")
)

type OrganizationService interface {
	Create(ctx context.Context, req *model.CreateOrganizationRequest) (*model.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateOrganizationRequest) (*model.Organization, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListOrganizationsRequest) ([]*model.Organization, int64, error)

	AddUser(ctx context.Context, orgID uuid.UUID, req *model.AddOrgUserRequest) (*model.OrganizationUser, error)
	UpdateUserRole(ctx context.Context, orgID, orgUserID uuid.UUID, req *model.UpdateOrgUserRoleRequest) (*model.OrganizationUser, error)
	RemoveUser(ctx context.Context, orgID, orgUserID uuid.UUID) error
	ListUsers(ctx context.Context, orgID uuid.UUID, req *model.ListOrgUsersRequest) ([]*model.OrganizationUser, int64, error)
}

type organizationService struct {
	orgRepo    repository.OrganizationRepository
	orgUserRepo repository.OrganizationUserRepository
	userRepo   repository.UserRepository
}

func NewOrganizationService(
	orgRepo repository.OrganizationRepository,
	orgUserRepo repository.OrganizationUserRepository,
	userRepo repository.UserRepository,
) OrganizationService {
	return &organizationService{
		orgRepo:    orgRepo,
		orgUserRepo: orgUserRepo,
		userRepo:   userRepo,
	}
}

func (s *organizationService) Create(ctx context.Context, req *model.CreateOrganizationRequest) (*model.Organization, error) {
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, ErrOwnerNotFound
	}

	exists, err := s.orgRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, apierror.Internal("Failed to check organization name")
	}
	if exists {
		return nil, ErrOrganizationExists
	}

	org := &model.Organization{
		Name:    req.Name,
		OwnerID: owner.ID,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, apierror.Internal("Failed to create organization")
	}

	orgUser := &model.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         owner.ID,
		Role:           model.OrgRoleAdmin,
	}

	if err := s.orgUserRepo.Create(ctx, orgUser); err != nil {
		return nil, apierror.Internal("Failed to add owner to organization")
	}

	return org, nil
}

func (s *organizationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	org, err := s.orgRepo.GetByIDWithOwner(ctx, id)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}
	return org, nil
}

func (s *organizationService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateOrganizationRequest) (*model.Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	if req.Name != "" {
		exists, err := s.orgRepo.ExistsByName(ctx, req.Name)
		if err != nil {
			return nil, apierror.Internal("Failed to check organization name")
		}
		if exists && org.Name != req.Name {
			return nil, ErrOrganizationExists
		}
		org.Name = req.Name
	}

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, apierror.Internal("Failed to update organization")
	}

	return s.orgRepo.GetByIDWithOwner(ctx, id)
}

func (s *organizationService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return ErrOrganizationNotFound
	}

	if err := s.orgRepo.Delete(ctx, id); err != nil {
		return apierror.Internal("Failed to delete organization")
	}

	return nil
}

func (s *organizationService) List(ctx context.Context, req *model.ListOrganizationsRequest) ([]*model.Organization, int64, error) {
	return s.orgRepo.List(ctx, req)
}

func (s *organizationService) AddUser(ctx context.Context, orgID uuid.UUID, req *model.AddOrgUserRequest) (*model.OrganizationUser, error) {
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	exists, err := s.orgUserRepo.ExistsByOrgAndUser(ctx, orgID, userID)
	if err != nil {
		return nil, apierror.Internal("Failed to check user membership")
	}
	if exists {
		return nil, ErrUserAlreadyInOrg
	}

	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           req.Role,
	}

	if err := s.orgUserRepo.Create(ctx, orgUser); err != nil {
		return nil, apierror.Internal("Failed to add user to organization")
	}

	return s.orgUserRepo.GetByID(ctx, orgUser.ID)
}

func (s *organizationService) UpdateUserRole(ctx context.Context, orgID, orgUserID uuid.UUID, req *model.UpdateOrgUserRoleRequest) (*model.OrganizationUser, error) {
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	orgUser, err := s.orgUserRepo.GetByID(ctx, orgUserID)
	if err != nil {
		return nil, ErrUserNotInOrg
	}

	if orgUser.OrganizationID != orgID {
		return nil, ErrUserNotInOrg
	}

	orgUser.Role = req.Role

	if err := s.orgUserRepo.Update(ctx, orgUser); err != nil {
		return nil, apierror.Internal("Failed to update user role")
	}

	return s.orgUserRepo.GetByID(ctx, orgUserID)
}

func (s *organizationService) RemoveUser(ctx context.Context, orgID, orgUserID uuid.UUID) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return ErrOrganizationNotFound
	}

	orgUser, err := s.orgUserRepo.GetByID(ctx, orgUserID)
	if err != nil {
		return ErrUserNotInOrg
	}

	if orgUser.OrganizationID != orgID {
		return ErrUserNotInOrg
	}

	if orgUser.UserID == org.OwnerID {
		return ErrCannotRemoveOwner
	}

	if err := s.orgUserRepo.Delete(ctx, orgUserID); err != nil {
		return apierror.Internal("Failed to remove user from organization")
	}

	return nil
}

func (s *organizationService) ListUsers(ctx context.Context, orgID uuid.UUID, req *model.ListOrgUsersRequest) ([]*model.OrganizationUser, int64, error) {
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, 0, ErrOrganizationNotFound
	}


	return s.orgUserRepo.ListByOrganization(ctx, orgID, req)
}
