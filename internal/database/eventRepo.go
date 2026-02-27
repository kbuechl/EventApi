package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Models ──────────────────────────────────────────────────────────────────

type RSVPStatus string

const (
	RSVPPending  RSVPStatus = "pending"
	RSVPGoing    RSVPStatus = "going"
	RSVPNotGoing RSVPStatus = "not_going"
)

type Event struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"       json:"id"`
	HostEmail   string         `gorm:"not null;index"             json:"hostEmail"`
	Name        string         `gorm:"not null"                   json:"name"`
	Description string         `json:"description"`
	Address     string         `json:"address"`
	Lat         float64        `json:"lat"`
	Long        float64        `json:"long"`
	StartTime   time.Time      `gorm:"not null"                   json:"startTime"`
	EndTime     time.Time      `gorm:"not null"                   json:"endTime"`
	AlbumID     string         `json:"albumId"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                      json:"-"`
}

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type EventInvitation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"  json:"id"`
	EventID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"eventId"`
	InviteeEmail string     `gorm:"not null;index"        json:"inviteeEmail"`
	RSVPStatus   RSVPStatus `gorm:"default:'pending'"     json:"rsvpStatus"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (e *EventInvitation) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type EventPhoto struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"     json:"id"`
	EventID       uuid.UUID `gorm:"type:uuid;not null;index" json:"eventId"`
	UploaderEmail string    `gorm:"not null"                 json:"uploaderEmail"`
	GooglePhotoID string    `gorm:"not null"                 json:"googlePhotoId"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (e *EventPhoto) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// ─── Repository ──────────────────────────────────────────────────────────────

type InvitationWithUser struct {
	EventInvitation
	FirstName string `gorm:"column:first_name" json:"firstName"`
	LastName  string `gorm:"column:last_name"  json:"lastName"`
}

type EventListRow struct {
	Event
	MyRSVP *RSVPStatus `gorm:"column:my_rsvp" json:"myRsvpStatus"`
}

type EventRepository interface {
	Create(event Event, inviteEmails []string) (*Event, error)
	GetByID(id uuid.UUID) (*Event, error)
	ListForUser(email string) ([]EventListRow, error)
	Update(id uuid.UUID, hostEmail string, fields map[string]interface{}) (*Event, error)
	Delete(id uuid.UUID, hostEmail string) error
	Invite(eventID uuid.UUID, hostEmail string, inviteEmails []string) error
	UpdateRSVP(eventID uuid.UUID, inviteeEmail string, status RSVPStatus) error
	GetInvitationsWithUsers(eventID uuid.UUID) ([]InvitationWithUser, error)
	GetPhoto(id uuid.UUID) (*EventPhoto, error)
	AddPhoto(photo EventPhoto) (*EventPhoto, error)
	DeletePhoto(id uuid.UUID, uploaderEmail string) error
	Migrate() error
}

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Migrate() error {
	return r.db.AutoMigrate(&Event{}, &EventInvitation{}, &EventPhoto{})
}

func (r *EventRepo) Create(event Event, inviteEmails []string) (*Event, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&event).Error; err != nil {
			return fmt.Errorf("error creating event: %w", err)
		}
		for _, email := range inviteEmails {
			inv := EventInvitation{EventID: event.ID, InviteeEmail: email}
			if err := tx.Create(&inv).Error; err != nil {
				return fmt.Errorf("error creating invitation for %s: %w", email, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepo) GetByID(id uuid.UUID) (*Event, error) {
	var event Event
	if err := r.db.First(&event, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepo) ListForUser(email string) ([]EventListRow, error) {
	var rows []EventListRow
	err := r.db.Table("events").
		Select("events.*, event_invitations.rsvp_status AS my_rsvp").
		Joins("LEFT JOIN event_invitations ON events.id = event_invitations.event_id AND event_invitations.invitee_email = ?", email).
		Where("events.deleted_at IS NULL AND (events.host_email = ? OR event_invitations.invitee_email = ?)", email, email).
		Order("events.start_time ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *EventRepo) Update(id uuid.UUID, hostEmail string, fields map[string]interface{}) (*Event, error) {
	result := r.db.Model(&Event{}).
		Where("id = ? AND host_email = ?", id, hostEmail).
		Updates(fields)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("event not found or permission denied")
	}
	return r.GetByID(id)
}

func (r *EventRepo) Delete(id uuid.UUID, hostEmail string) error {
	result := r.db.Where("id = ? AND host_email = ?", id, hostEmail).Delete(&Event{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found or permission denied")
	}
	return nil
}

func (r *EventRepo) Invite(eventID uuid.UUID, hostEmail string, inviteEmails []string) error {
	var event Event
	if err := r.db.First(&event, "id = ? AND host_email = ?", eventID, hostEmail).Error; err != nil {
		return fmt.Errorf("event not found or permission denied")
	}
	for _, email := range inviteEmails {
		inv := EventInvitation{EventID: eventID, InviteeEmail: email}
		// Use Clauses to ignore duplicate invitations
		if err := r.db.Where(EventInvitation{EventID: eventID, InviteeEmail: email}).
			FirstOrCreate(&inv).Error; err != nil {
			return fmt.Errorf("error inviting %s: %w", email, err)
		}
	}
	return nil
}

func (r *EventRepo) UpdateRSVP(eventID uuid.UUID, inviteeEmail string, status RSVPStatus) error {
	result := r.db.Model(&EventInvitation{}).
		Where("event_id = ? AND invitee_email = ?", eventID, inviteeEmail).
		Update("rsvp_status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("invitation not found")
	}
	return nil
}

func (r *EventRepo) GetInvitationsWithUsers(eventID uuid.UUID) ([]InvitationWithUser, error) {
	var invitations []InvitationWithUser
	err := r.db.Table("event_invitations").
		Select("event_invitations.*, users.first_name, users.last_name").
		Joins("JOIN users ON users.email = event_invitations.invitee_email").
		Where("event_invitations.event_id = ?", eventID).
		Scan(&invitations).Error
	return invitations, err
}

func (r *EventRepo) GetPhoto(id uuid.UUID) (*EventPhoto, error) {
	var photo EventPhoto
	if err := r.db.First(&photo, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &photo, nil
}

func (r *EventRepo) AddPhoto(photo EventPhoto) (*EventPhoto, error) {
	if err := r.db.Create(&photo).Error; err != nil {
		return nil, fmt.Errorf("error adding photo: %w", err)
	}
	return &photo, nil
}

func (r *EventRepo) DeletePhoto(id uuid.UUID, uploaderEmail string) error {
	result := r.db.Where("id = ? AND uploader_email = ?", id, uploaderEmail).Delete(&EventPhoto{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("photo not found or permission denied")
	}
	return nil
}
