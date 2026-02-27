package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ─── Model ───────────────────────────────────────────────────────────────────

type Contact struct {
	OwnerEmail   string    `gorm:"primaryKey" json:"ownerEmail"`
	ContactEmail string    `gorm:"primaryKey" json:"contactEmail"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ContactWithUser enriches a Contact with the contact user's name and picture.
type ContactWithUser struct {
	Contact
	FirstName string `gorm:"column:first_name" json:"firstName"`
	LastName  string `gorm:"column:last_name"  json:"lastName"`
	Picture   string `gorm:"column:picture"    json:"picture"`
}

// ─── Repository ──────────────────────────────────────────────────────────────

type ContactRepository interface {
	List(ownerEmail string) ([]ContactWithUser, error)
	Add(ownerEmail, contactEmail string) (*Contact, error)
	Remove(ownerEmail, contactEmail string) error
	Migrate() error
}

type ContactRepo struct {
	db *gorm.DB
}

func NewContactRepo(db *gorm.DB) *ContactRepo {
	return &ContactRepo{db: db}
}

func (r *ContactRepo) Migrate() error {
	return r.db.AutoMigrate(&Contact{})
}

func (r *ContactRepo) List(ownerEmail string) ([]ContactWithUser, error) {
	var contacts []ContactWithUser
	err := r.db.Table("contacts").
		Select("contacts.*, users.first_name, users.last_name, users.picture").
		Joins("JOIN users ON users.email = contacts.contact_email").
		Where("contacts.owner_email = ?", ownerEmail).
		Order("users.first_name ASC").
		Scan(&contacts).Error
	return contacts, err
}

func (r *ContactRepo) Add(ownerEmail, contactEmail string) (*Contact, error) {
	if ownerEmail == contactEmail {
		return nil, fmt.Errorf("cannot add yourself as a contact")
	}
	contact := Contact{OwnerEmail: ownerEmail, ContactEmail: contactEmail}
	result := r.db.Where(Contact{OwnerEmail: ownerEmail, ContactEmail: contactEmail}).
		FirstOrCreate(&contact)
	if result.Error != nil {
		return nil, fmt.Errorf("error adding contact: %w", result.Error)
	}
	return &contact, nil
}

func (r *ContactRepo) Remove(ownerEmail, contactEmail string) error {
	result := r.db.Where("owner_email = ? AND contact_email = ?", ownerEmail, contactEmail).
		Delete(&Contact{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("contact not found")
	}
	return nil
}
