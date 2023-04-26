package database

import (
	"time"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

type UserRepository interface {
	Create(u NewUser) (*User, error)
	Exists(id string) bool
	Migrate() error
}

type NewUser struct {
	Email        string `gorm:"primaryKey"`
	ID           string
	FirstName    string
	LastName     string
	Picture      string
	Verified     bool
	RefreshToken string
}

type User struct {
	NewUser
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (s *UserRepo) Migrate() error {
	return s.db.AutoMigrate(&User{})
}

func (s *UserRepo) Create(u NewUser) (*User, error) {
	newUser := User{
		NewUser: u,
	}

	result := s.db.Create(&newUser)

	if result.Error != nil {
		return nil, result.Error
	}

	return &newUser, nil
}

func (s *UserRepo) Exists(id string) bool {
	var user User
	result := s.db.First(&user, "id=?", id)

	return result.RowsAffected > 0
}
