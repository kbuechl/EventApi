package database

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

type UserRepository interface {
	Create(u User) (*User, error)
	Update(email string, u *OidcUser, refreshToken string) (*User, error)
	Get(email string) (*User, error)
	Migrate() error
}

type OidcUser struct {
	Email     string `json:"email"`
	Verified  bool   `json:"email_verified"`
	Picture   string `json:"picture"`
	LastName  string `json:"family_name"`
	FirstName string `json:"given_name"`
}

type User struct {
	Email        string `gorm:"primaryKey"`
	FirstName    string
	LastName     string
	Picture      string
	Verified     bool
	RefreshToken string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    time.Time
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (s *UserRepo) Migrate() error {
	return s.db.AutoMigrate(&User{})
}

func (s *UserRepo) Create(u User) (*User, error) {
	result := s.db.Create(&u)

	if result.Error != nil {
		return nil, fmt.Errorf("error creating user: %w", result.Error)
	}

	return &u, nil
}

func (s *UserRepo) Update(email string, u *OidcUser, refreshToken string) (*User, error) {
	user := &User{
		Email: email,
	}
	result := s.db.First(user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("error updating user: user does not exist")
		}
		return nil, fmt.Errorf("error updating user: %w", result.Error)
	}

	user.Picture = u.Picture

	//only sent on first auth
	if refreshToken != "" {
		user.RefreshToken = refreshToken
	}

	user.FirstName = u.FirstName
	user.LastName = u.LastName
	user.Verified = u.Verified

	s.db.Save(user)
	return user, nil
}

func (s *UserRepo) Get(email string) (*User, error) {
	var user User
	result := s.db.First(&user, "email=?", email)
	return &user, result.Error
}
