package service

import (
	"context"
	"errors"
	"testing"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ronaldocristover/lms-backend/internal/model"
)

type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetByIDWithOwner(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) Update(ctx context.Context, org *model.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationRepository) List(ctx context.Context, filter *model.ListOrganizationsRequest) ([]*model.Organization, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Organization), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrganizationRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(bool), args.Error(1)
}

type MockOrganizationUserRepository struct {
	mock.Mock
}

func (m *MockOrganizationUserRepository) Create(ctx context.Context, orgUser *model.OrganizationUser) error {
	args := m.Called(ctx, orgUser)
	return args.Error(0)
}

func (m *MockOrganizationUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.OrganizationUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationUser), args.Error(1)
}

func (m *MockOrganizationUserRepository) GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*model.OrganizationUser, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationUser), args.Error(1)
}

func (m *MockOrganizationUserRepository) Update(ctx context.Context, orgUser *model.OrganizationUser) error {
	args := m.Called(ctx, orgUser)
	return args.Error(0)
}

func (m *MockOrganizationUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationUserRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, filter *model.ListOrgUsersRequest) ([]*model.OrganizationUser, int64, error) {
	args := m.Called(ctx, orgID, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.OrganizationUser), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrganizationUserRepository) ExistsByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockOrganizationUserRepository) DeleteByOrganization(ctx context.Context, orgID uuid.UUID) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) List(ctx context.Context, filter *model.ListUsersRequest) ([]*model.User, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

func newTestOrgService() (OrganizationService, *MockOrganizationRepository, *MockOrganizationUserRepository, *MockUserRepo) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockOrgUserRepo := new(MockOrganizationUserRepository)
	mockUserRepo := new(MockUserRepo)
	svc := NewOrganizationService(mockOrgRepo, mockOrgUserRepo, mockUserRepo, zap.NewNop().Sugar())
	return svc, mockOrgRepo, mockOrgUserRepo, mockUserRepo
}

// ─── CREATE ───

func TestOrganizationService_Create_Success(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, mockUserRepo := newTestOrgService()

	ownerID := uuid.New()
	owner := &model.User{ID: ownerID, Email: "owner@example.com", Name: "Owner"}

	mockUserRepo.On("GetByID", mock.Anything, ownerID).Return(owner, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "Test Org").Return(false, nil)
	mockOrgRepo.On("Create", mock.Anything, mock.MatchedBy(func(org *model.Organization) bool {
		return org.Name == "Test Org" && org.OwnerID == ownerID
	})).Return(nil)
	mockOrgUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(ou *model.OrganizationUser) bool {
		return ou.UserID == ownerID && ou.Role == model.OrgRoleAdmin
	})).Return(nil)

	org, err := svc.Create(context.Background(), &model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, "Test Org", org.Name)
	assert.Equal(t, ownerID, org.OwnerID)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_Create_InvalidOwnerID(t *testing.T) {
	svc, _, _, _ := newTestOrgService()

	org, err := svc.Create(context.Background(), &model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: "invalid-uuid",
	})

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidUserID, err)
	assert.Nil(t, org)
}

func TestOrganizationService_Create_OwnerNotFound(t *testing.T) {
	svc, _, _, mockUserRepo := newTestOrgService()

	ownerID := uuid.New()
	mockUserRepo.On("GetByID", mock.Anything, ownerID).Return(nil, errors.New("not found"))

	org, err := svc.Create(context.Background(), &model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})

	assert.Error(t, err)
	assert.Equal(t, ErrOwnerNotFound, err)
	assert.Nil(t, org)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_Create_NameExists(t *testing.T) {
	svc, mockOrgRepo, _, mockUserRepo := newTestOrgService()

	ownerID := uuid.New()
	owner := &model.User{ID: ownerID, Email: "owner@example.com", Name: "Owner"}

	mockUserRepo.On("GetByID", mock.Anything, ownerID).Return(owner, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "Existing Org").Return(true, nil)

	org, err := svc.Create(context.Background(), &model.CreateOrganizationRequest{
		Name:    "Existing Org",
		OwnerID: ownerID.String(),
	})

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationExists, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_Create_RepoError(t *testing.T) {
	svc, mockOrgRepo, _, mockUserRepo := newTestOrgService()

	ownerID := uuid.New()
	owner := &model.User{ID: ownerID, Email: "owner@example.com", Name: "Owner"}

	mockUserRepo.On("GetByID", mock.Anything, ownerID).Return(owner, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "Test Org").Return(false, nil)
	mockOrgRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	org, err := svc.Create(context.Background(), &model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})

	assert.Error(t, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_Create_ExistsByNameError(t *testing.T) {
	svc, mockOrgRepo, _, mockUserRepo := newTestOrgService()

	ownerID := uuid.New()
	owner := &model.User{ID: ownerID, Email: "owner@example.com", Name: "Owner"}

	mockUserRepo.On("GetByID", mock.Anything, ownerID).Return(owner, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "Test Org").Return(false, errors.New("db error"))

	org, err := svc.Create(context.Background(), &model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})

	assert.Error(t, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_Create_OrgUserCreateError(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, mockUserRepo := newTestOrgService()

	ownerID := uuid.New()
	owner := &model.User{ID: ownerID, Email: "owner@example.com", Name: "Owner"}

	mockUserRepo.On("GetByID", mock.Anything, ownerID).Return(owner, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "Test Org").Return(false, nil)
	mockOrgRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockOrgUserRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	org, err := svc.Create(context.Background(), &model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})

	assert.Error(t, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// ─── GET BY ID ───

func TestOrganizationService_GetByID_Success(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	expected := &model.Organization{ID: orgID, Name: "Test Org", OwnerID: uuid.New()}

	mockOrgRepo.On("GetByIDWithOwner", mock.Anything, orgID).Return(expected, nil)

	org, err := svc.GetByID(context.Background(), orgID)

	assert.NoError(t, err)
	assert.Equal(t, expected, org)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_GetByID_NotFound(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByIDWithOwner", mock.Anything, orgID).Return(nil, errors.New("not found"))

	org, err := svc.GetByID(context.Background(), orgID)

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationNotFound, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
}

// ─── UPDATE ───

func TestOrganizationService_Update_Success(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	existing := &model.Organization{ID: orgID, Name: "Old Name", OwnerID: uuid.New()}
	updated := &model.Organization{ID: orgID, Name: "New Name", OwnerID: existing.OwnerID}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(existing, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "New Name").Return(false, nil)
	mockOrgRepo.On("Update", mock.Anything, mock.MatchedBy(func(org *model.Organization) bool {
		return org.Name == "New Name"
	})).Return(nil)
	mockOrgRepo.On("GetByIDWithOwner", mock.Anything, orgID).Return(updated, nil)

	org, err := svc.Update(context.Background(), orgID, &model.UpdateOrganizationRequest{
		Name: "New Name",
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", org.Name)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_Update_NotFound(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(nil, errors.New("not found"))

	org, err := svc.Update(context.Background(), orgID, &model.UpdateOrganizationRequest{
		Name: "New Name",
	})

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationNotFound, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_Update_NameExists(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	existing := &model.Organization{ID: orgID, Name: "Old Name", OwnerID: uuid.New()}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(existing, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "Existing Name").Return(true, nil)

	org, err := svc.Update(context.Background(), orgID, &model.UpdateOrganizationRequest{
		Name: "Existing Name",
	})

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationExists, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_Update_SameName(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	existing := &model.Organization{ID: orgID, Name: "Same Name", OwnerID: uuid.New()}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(existing, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "Same Name").Return(true, nil)
	mockOrgRepo.On("Update", mock.Anything, existing).Return(nil)
	mockOrgRepo.On("GetByIDWithOwner", mock.Anything, orgID).Return(existing, nil)

	org, err := svc.Update(context.Background(), orgID, &model.UpdateOrganizationRequest{
		Name: "Same Name",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Same Name", org.Name)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_Update_RepoError(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	existing := &model.Organization{ID: orgID, Name: "Old Name", OwnerID: uuid.New()}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(existing, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "New Name").Return(false, nil)
	mockOrgRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	org, err := svc.Update(context.Background(), orgID, &model.UpdateOrganizationRequest{
		Name: "New Name",
	})

	assert.Error(t, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_Update_ExistsByNameError(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	existing := &model.Organization{ID: orgID, Name: "Old Name", OwnerID: uuid.New()}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(existing, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "New Name").Return(false, errors.New("db error"))

	org, err := svc.Update(context.Background(), orgID, &model.UpdateOrganizationRequest{
		Name: "New Name",
	})

	assert.Error(t, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_Update_GetByIDWithOwnerError(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	existing := &model.Organization{ID: orgID, Name: "Old Name", OwnerID: uuid.New()}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(existing, nil)
	mockOrgRepo.On("ExistsByName", mock.Anything, "New Name").Return(false, nil)
	mockOrgRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockOrgRepo.On("GetByIDWithOwner", mock.Anything, orgID).Return(nil, errors.New("db error"))

	org, err := svc.Update(context.Background(), orgID, &model.UpdateOrganizationRequest{
		Name: "New Name",
	})

	assert.Error(t, err)
	assert.Nil(t, org)
	mockOrgRepo.AssertExpectations(t)
}

// ─── DELETE ───

func TestOrganizationService_Delete_Success(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgRepo.On("Delete", mock.Anything, orgID).Return(nil)

	err := svc.Delete(context.Background(), orgID)

	assert.NoError(t, err)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_Delete_NotFound(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(nil, errors.New("not found"))

	err := svc.Delete(context.Background(), orgID)

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationNotFound, err)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_Delete_RepoError(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgRepo.On("Delete", mock.Anything, orgID).Return(errors.New("db error"))

	err := svc.Delete(context.Background(), orgID)

	assert.Error(t, err)
	mockOrgRepo.AssertExpectations(t)
}

// ─── LIST ───

func TestOrganizationService_List_Success(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgs := []*model.Organization{
		{ID: uuid.New(), Name: "Org 1", OwnerID: uuid.New()},
		{ID: uuid.New(), Name: "Org 2", OwnerID: uuid.New()},
	}

	mockOrgRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListOrganizationsRequest) bool {
		return true // pagination handled by repository
	})).Return(orgs, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListOrganizationsRequest{
		Page:     1,
		PageSize: 20,
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_List_DefaultPagination(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	mockOrgRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListOrganizationsRequest) bool {
		return true // pagination handled by repository
	})).Return([]*model.Organization{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListOrganizationsRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_List_WithSearch(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	mockOrgRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListOrganizationsRequest) bool {
		return req.Search == "test" && req.Page == 0 && req.PageSize == 0
	})).Return([]*model.Organization{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListOrganizationsRequest{
		Search: "test",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_List_WithOwnerFilter(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	ownerID := uuid.New()
	mockOrgRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListOrganizationsRequest) bool {
		return req.OwnerID == ownerID.String()
	})).Return([]*model.Organization{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListOrganizationsRequest{
		OwnerID: ownerID.String(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_List_WithUserIDFilter(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	userID := uuid.New()
	mockOrgRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListOrganizationsRequest) bool {
		return req.UserID == userID.String()
	})).Return([]*model.Organization{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListOrganizationsRequest{
		UserID: userID.String(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockOrgRepo.AssertExpectations(t)
}

// ─── ADD USER ───

func TestOrganizationService_AddUser_Success(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, mockUserRepo := newTestOrgService()

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID}, nil)
	mockOrgUserRepo.On("ExistsByOrgAndUser", mock.Anything, orgID, userID).Return(false, nil)
	mockOrgUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, mock.Anything).Return(orgUser, nil)

	result, err := svc.AddUser(context.Background(), orgID, &model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.OrgRoleMember, result.Role)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_AddUser_OrgNotFound(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(nil, errors.New("not found"))

	result, err := svc.AddUser(context.Background(), orgID, &model.AddOrgUserRequest{
		UserID: uuid.New().String(),
		Role:   model.OrgRoleMember,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationNotFound, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_AddUser_InvalidUserID(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)

	result, err := svc.AddUser(context.Background(), orgID, &model.AddOrgUserRequest{
		UserID: "invalid-uuid",
		Role:   model.OrgRoleMember,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidUserID, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_AddUser_UserNotFound(t *testing.T) {
	svc, mockOrgRepo, _, mockUserRepo := newTestOrgService()

	orgID := uuid.New()
	userID := uuid.New()

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, errors.New("not found"))

	result, err := svc.AddUser(context.Background(), orgID, &model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_AddUser_AlreadyInOrg(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, mockUserRepo := newTestOrgService()

	orgID := uuid.New()
	userID := uuid.New()

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID}, nil)
	mockOrgUserRepo.On("ExistsByOrgAndUser", mock.Anything, orgID, userID).Return(true, nil)

	result, err := svc.AddUser(context.Background(), orgID, &model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrUserAlreadyInOrg, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_AddUser_ExistsByOrgAndUserError(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, mockUserRepo := newTestOrgService()

	orgID := uuid.New()
	userID := uuid.New()

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID}, nil)
	mockOrgUserRepo.On("ExistsByOrgAndUser", mock.Anything, orgID, userID).Return(false, errors.New("db error"))

	result, err := svc.AddUser(context.Background(), orgID, &model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestOrganizationService_AddUser_CreateError(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, mockUserRepo := newTestOrgService()

	orgID := uuid.New()
	userID := uuid.New()

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID}, nil)
	mockOrgUserRepo.On("ExistsByOrgAndUser", mock.Anything, orgID, userID).Return(false, nil)
	mockOrgUserRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	result, err := svc.AddUser(context.Background(), orgID, &model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// ─── UPDATE USER ROLE ───

func TestOrganizationService_UpdateUserRole_Success(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	orgUserID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             orgUserID,
		OrganizationID: orgID,
		UserID:         uuid.New(),
		Role:           model.OrgRoleMember,
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(orgUser, nil)
	mockOrgUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(ou *model.OrganizationUser) bool {
		return ou.Role == model.OrgRoleAdmin
	})).Return(nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(&model.OrganizationUser{
		ID:             orgUserID,
		OrganizationID: orgID,
		UserID:         orgUser.UserID,
		Role:           model.OrgRoleAdmin,
	}, nil)

	result, err := svc.UpdateUserRole(context.Background(), orgID, orgUserID, &model.UpdateOrgUserRoleRequest{
		Role: model.OrgRoleAdmin,
	})

	assert.NoError(t, err)
	assert.Equal(t, model.OrgRoleAdmin, result.Role)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_UpdateUserRole_OrgNotFound(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(nil, errors.New("not found"))

	result, err := svc.UpdateUserRole(context.Background(), orgID, uuid.New(), &model.UpdateOrgUserRoleRequest{
		Role: model.OrgRoleAdmin,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationNotFound, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_UpdateUserRole_UserNotInOrg(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	orgUserID := uuid.New()

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(nil, errors.New("not found"))

	result, err := svc.UpdateUserRole(context.Background(), orgID, orgUserID, &model.UpdateOrgUserRoleRequest{
		Role: model.OrgRoleAdmin,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotInOrg, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_UpdateUserRole_WrongOrg(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	wrongOrgID := uuid.New()
	orgUserID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             orgUserID,
		OrganizationID: wrongOrgID,
		UserID:         uuid.New(),
		Role:           model.OrgRoleMember,
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(orgUser, nil)

	result, err := svc.UpdateUserRole(context.Background(), orgID, orgUserID, &model.UpdateOrgUserRoleRequest{
		Role: model.OrgRoleAdmin,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotInOrg, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_UpdateUserRole_UpdateError(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	orgUserID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             orgUserID,
		OrganizationID: orgID,
		UserID:         uuid.New(),
		Role:           model.OrgRoleMember,
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(orgUser, nil)
	mockOrgUserRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	result, err := svc.UpdateUserRole(context.Background(), orgID, orgUserID, &model.UpdateOrgUserRoleRequest{
		Role: model.OrgRoleAdmin,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

// ─── REMOVE USER ───

func TestOrganizationService_RemoveUser_Success(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	orgUserID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             orgUserID,
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID, OwnerID: uuid.New()}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(orgUser, nil)
	mockOrgUserRepo.On("Delete", mock.Anything, orgUserID).Return(nil)

	err := svc.RemoveUser(context.Background(), orgID, orgUserID)

	assert.NoError(t, err)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_RemoveUser_OrgNotFound(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(nil, errors.New("not found"))

	err := svc.RemoveUser(context.Background(), orgID, uuid.New())

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationNotFound, err)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_RemoveUser_UserNotInOrg(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	orgUserID := uuid.New()

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(nil, errors.New("not found"))

	err := svc.RemoveUser(context.Background(), orgID, orgUserID)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotInOrg, err)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_RemoveUser_CannotRemoveOwner(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	ownerID := uuid.New()
	orgUserID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             orgUserID,
		OrganizationID: orgID,
		UserID:         ownerID,
		Role:           model.OrgRoleAdmin,
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID, OwnerID: ownerID}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(orgUser, nil)

	err := svc.RemoveUser(context.Background(), orgID, orgUserID)

	assert.Error(t, err)
	assert.Equal(t, ErrCannotRemoveOwner, err)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_RemoveUser_WrongOrg(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	wrongOrgID := uuid.New()
	orgUserID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             orgUserID,
		OrganizationID: wrongOrgID,
		UserID:         uuid.New(),
		Role:           model.OrgRoleMember,
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID, OwnerID: uuid.New()}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(orgUser, nil)

	err := svc.RemoveUser(context.Background(), orgID, orgUserID)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotInOrg, err)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_RemoveUser_DeleteError(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	orgUserID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             orgUserID,
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID, OwnerID: uuid.New()}, nil)
	mockOrgUserRepo.On("GetByID", mock.Anything, orgUserID).Return(orgUser, nil)
	mockOrgUserRepo.On("Delete", mock.Anything, orgUserID).Return(errors.New("db error"))

	err := svc.RemoveUser(context.Background(), orgID, orgUserID)

	assert.Error(t, err)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

// ─── LIST USERS ───

func TestOrganizationService_ListUsers_Success(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	users := []*model.OrganizationUser{
		{ID: uuid.New(), OrganizationID: orgID, UserID: uuid.New(), Role: model.OrgRoleAdmin},
		{ID: uuid.New(), OrganizationID: orgID, UserID: uuid.New(), Role: model.OrgRoleMember},
	}

	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("ListByOrganization", mock.Anything, orgID, mock.MatchedBy(func(req *model.ListOrgUsersRequest) bool {
		return true // pagination handled by repository
	})).Return(users, int64(2), nil)

	result, total, err := svc.ListUsers(context.Background(), orgID, &model.ListOrgUsersRequest{
		Page:     1,
		PageSize: 20,
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_ListUsers_OrgNotFound(t *testing.T) {
	svc, mockOrgRepo, _, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(nil, errors.New("not found"))

	result, total, err := svc.ListUsers(context.Background(), orgID, &model.ListOrgUsersRequest{})

	assert.Error(t, err)
	assert.Equal(t, ErrOrganizationNotFound, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, result)
	mockOrgRepo.AssertExpectations(t)
}

func TestOrganizationService_ListUsers_DefaultPagination(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("ListByOrganization", mock.Anything, orgID, mock.MatchedBy(func(req *model.ListOrgUsersRequest) bool {
		return true // pagination handled by repository
	})).Return([]*model.OrganizationUser{}, int64(0), nil)

	result, _, err := svc.ListUsers(context.Background(), orgID, &model.ListOrgUsersRequest{})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_ListUsers_WithRoleFilter(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("ListByOrganization", mock.Anything, orgID, mock.MatchedBy(func(req *model.ListOrgUsersRequest) bool {
		return req.Role == model.OrgRoleAdmin
	})).Return([]*model.OrganizationUser{}, int64(0), nil)

	result, _, err := svc.ListUsers(context.Background(), orgID, &model.ListOrgUsersRequest{
		Role: model.OrgRoleAdmin,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}

func TestOrganizationService_ListUsers_WithSearch(t *testing.T) {
	svc, mockOrgRepo, mockOrgUserRepo, _ := newTestOrgService()

	orgID := uuid.New()
	mockOrgRepo.On("GetByID", mock.Anything, orgID).Return(&model.Organization{ID: orgID}, nil)
	mockOrgUserRepo.On("ListByOrganization", mock.Anything, orgID, mock.MatchedBy(func(req *model.ListOrgUsersRequest) bool {
		return req.Search == "test"
	})).Return([]*model.OrganizationUser{}, int64(0), nil)

	result, _, err := svc.ListUsers(context.Background(), orgID, &model.ListOrgUsersRequest{
		Search: "test",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockOrgRepo.AssertExpectations(t)
	mockOrgUserRepo.AssertExpectations(t)
}
