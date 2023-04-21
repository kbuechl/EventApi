package database

import (
	"time"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

type NewUser struct {
	ID           string `gorm:"primaryKey"`
	FirstName    string
	LastName     string
	Email        string
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

func NewUserService() *UserService {
	return &UserService{
		DB: NewClient(),
	}
}

func init() {
	//todo: can we run this as a separate script at controlled points?
	db := NewClient()
	db.AutoMigrate(&User{})
}

func (s *UserService) Create(u NewUser) (*User, error) {
	newUser := User{
		NewUser: u,
	}

	result := s.DB.Create(&newUser)

	if result.Error != nil {
		return nil, result.Error
	}

	return &newUser, nil
}

func (s *UserService) Exists(id string) bool {
	var user User
	result := s.DB.First(&user, "id=?", id)

	return result.RowsAffected > 0
}
