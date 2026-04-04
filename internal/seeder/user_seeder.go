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
	Role     string
	Status   string
	Avatar   string
}

func (s *UserSeeder) Seed(ctx context.Context) error {
	seeds := []SeedUser{
		{
			Email:    "admin@lms.com",
			Password: "admin12345",
			Name:     "System Admin",
			Role:     model.UserRoleAdmin,
			Status:   model.UserStatusActive,
			Avatar:   "https://api.dicebear.com/7.x/avataaars/svg?seed=admin",
		},
		{
			Email:    "tutor@lms.com",
			Password: "tutor12345",
			Name:     "John Tutor",
			Role:     model.UserRoleTutor,
			Status:   model.UserStatusActive,
			Avatar:   "https://api.dicebear.com/7.x/avataaars/svg?seed=tutor",
		},
		{
			Email:    "student@lms.com",
			Password: "student12345",
			Name:     "Jane Student",
			Role:     model.UserRoleUser,
			Status:   model.UserStatusActive,
			Avatar:   "https://api.dicebear.com/7.x/avataaars/svg?seed=student",
		},
		{
			Email:    "suspended@lms.com",
			Password: "suspended123",
			Name:     "Suspended User",
			Role:     model.UserRoleUser,
			Status:   model.UserStatusSuspended,
			Avatar:   "",
		},
		{
			Email:    "inactive@lms.com",
			Password: "inactive12345",
			Name:     "Inactive User",
			Role:     model.UserRoleUser,
			Status:   model.UserStatusInactive,
			Avatar:   "",
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

		user := model.User{
			ID:           uuid.New(),
			Email:        seed.Email,
			PasswordHash: string(hashedPassword),
			Name:         seed.Name,
			Role:         seed.Role,
			Status:       seed.Status,
			Avatar:       seed.Avatar,
		}

		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return fmt.Errorf("failed to seed user %s: %w", seed.Email, err)
		}

		log.Printf("✅ Seeded %s (%s)", seed.Email, seed.Role)
	}

	return nil
}

func (s *UserSeeder) Truncate(ctx context.Context) error {
	return s.db.WithContext(ctx).Exec("DELETE FROM users").Error
}
