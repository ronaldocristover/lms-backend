package seeder

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/yourusername/lms/internal/model"
)

type UserSeeder struct {
	db *gorm.DB
}

func NewUserSeeder(db *gorm.DB) *UserSeeder {
	return &UserSeeder{db: db}
}

type SeedUser struct {
	Email    string
	Password string
	Name     string
	RoleName string
}

func (s *UserSeeder) Seed(ctx context.Context) error {
	roles := map[string]*model.Role{}
	for _, roleName := range []string{model.RoleAdmin, model.RoleStudent, model.RoleTutor, model.RoleOrgAdmin} {
		role, err := s.getOrCreateRole(ctx, roleName)
		if err != nil {
			return fmt.Errorf("failed to seed role %s: %w", roleName, err)
		}
		roles[roleName] = role
	}

	seeds := []SeedUser{
		{
			Email:    "admin@lms.com",
			Password: "admin12345",
			Name:     "System Admin",
			RoleName: model.RoleAdmin,
		},
		{
			Email:    "tutor@lms.com",
			Password: "tutor12345",
			Name:     "John Tutor",
			RoleName: model.RoleTutor,
		},
		{
			Email:    "student@lms.com",
			Password: "student12345",
			Name:     "Jane Student",
			RoleName: model.RoleStudent,
		},
		{
			Email:    "orgadmin@lms.com",
			Password: "orgadmin12345",
			Name:     "Org Admin",
			RoleName: model.RoleOrgAdmin,
		},
	}

	for _, seed := range seeds {
		var existing model.User
		err := s.db.WithContext(ctx).Where("email = ?", seed.Email).First(&existing).Error
		if err == nil {
			log.Printf("⏭️  Skipping %s (already exists)", seed.Email)
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", seed.Email, err)
		}

		role := roles[seed.RoleName]
		user := model.User{
			ID:           uuid.New(),
			Email:        seed.Email,
			PasswordHash: string(hashedPassword),
			Name:         seed.Name,
			RoleID:       role.ID,
		}

		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return fmt.Errorf("failed to seed user %s: %w", seed.Email, err)
		}

		log.Printf("✅ Seeded %s (%s)", seed.Email, seed.RoleName)
	}

	return nil
}

func (s *UserSeeder) getOrCreateRole(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := s.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err == nil {
		return &role, nil
	}

	role = model.Role{
		ID:   uuid.New(),
		Name: name,
	}
	if err := s.db.WithContext(ctx).Create(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (s *UserSeeder) Truncate(ctx context.Context) error {
	if err := s.db.WithContext(ctx).Exec("DELETE FROM users").Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Exec("DELETE FROM roles").Error
}
