package events

import (
	"eventapi/internal/database"
	"eventapi/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Request / Response types ─────────────────────────────────────────────────

type createEventRequest struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Address      string    `json:"address"`
	Lat          float64   `json:"lat"`
	Long         float64   `json:"long"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	InviteEmails []string  `json:"inviteEmails"`
}

type updateEventRequest struct {
	Address   string    `json:"address"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type inviteRequest struct {
	Emails []string `json:"emails"`
}

type rsvpRequest struct {
	Status database.RSVPStatus `json:"status"`
}

type eventListItem struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Address     string               `json:"address"`
	StartTime   time.Time            `json:"startTime"`
	EndTime     time.Time            `json:"endTime"`
	HostEmail   string               `json:"hostEmail"`
	IsHost      bool                 `json:"isHost"`
	MyRSVP      *database.RSVPStatus `json:"myRsvpStatus"`
}

type invitationDetail struct {
	InviteeEmail string              `json:"inviteeEmail"`
	FirstName    string              `json:"firstName"`
	LastName     string              `json:"lastName"`
	RSVPStatus   database.RSVPStatus `json:"rsvpStatus"`
}

type eventDetail struct {
	eventListItem
	AlbumID     string             `json:"albumId"`
	Lat         float64            `json:"lat"`
	Long        float64            `json:"long"`
	Invitations []invitationDetail `json:"invitations"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

type Controller struct {
	repo     database.EventRepository
	userRepo database.UserRepository
}

func NewController(repo database.EventRepository, userRepo database.UserRepository) *Controller {
	return &Controller{repo: repo, userRepo: userRepo}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func parseEventID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid event ID")
	}
	return id, nil
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// List returns all events the current user hosts or is invited to.
func (ctrl *Controller) List(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)
	rows, err := ctrl.repo.ListForUser(sd.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	items := make([]eventListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, eventListItem{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Address:     r.Address,
			StartTime:   r.StartTime,
			EndTime:     r.EndTime,
			HostEmail:   r.HostEmail,
			IsHost:      r.HostEmail == sd.Email,
			MyRSVP:      r.MyRSVP,
		})
	}
	return c.JSON(items)
}

// Create creates a new event with optional invitees.
func (ctrl *Controller) Create(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)

	var req createEventRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		return fiber.NewError(fiber.StatusBadRequest, "startTime and endTime are required")
	}

	event := database.Event{
		HostEmail:   sd.Email,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Lat:         req.Lat,
		Long:        req.Long,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		// AlbumID set in Phase 6 when Google Photos album is created
	}

	created, err := ctrl.repo.Create(event, req.InviteEmails)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// Get returns full detail for a single event.
func (ctrl *Controller) Get(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)
	id, err := parseEventID(c)
	if err != nil {
		return err
	}

	event, err := ctrl.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	invitations, err := ctrl.repo.GetInvitationsWithUsers(id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	isHost := event.HostEmail == sd.Email

	// Find current user's RSVP
	var myRSVP *database.RSVPStatus
	details := make([]invitationDetail, 0, len(invitations))
	for _, inv := range invitations {
		if inv.InviteeEmail == sd.Email {
			status := inv.RSVPStatus
			myRSVP = &status
		}
		// Guests only see who is going; hosts see all statuses
		if isHost || inv.RSVPStatus == database.RSVPGoing {
			details = append(details, invitationDetail{
				InviteeEmail: inv.InviteeEmail,
				FirstName:    inv.FirstName,
				LastName:     inv.LastName,
				RSVPStatus:   inv.RSVPStatus,
			})
		}
	}

	return c.JSON(eventDetail{
		eventListItem: eventListItem{
			ID:          event.ID,
			Name:        event.Name,
			Description: event.Description,
			Address:     event.Address,
			StartTime:   event.StartTime,
			EndTime:     event.EndTime,
			HostEmail:   event.HostEmail,
			IsHost:      isHost,
			MyRSVP:      myRSVP,
		},
		AlbumID:     event.AlbumID,
		Lat:         event.Lat,
		Long:        event.Long,
		Invitations: details,
	})
}

// Update allows the host to change event address or times.
func (ctrl *Controller) Update(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)
	id, err := parseEventID(c)
	if err != nil {
		return err
	}

	var req updateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	fields := map[string]interface{}{
		"address":    req.Address,
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
	}
	updated, err := ctrl.repo.Update(id, sd.Email, fields)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}
	return c.JSON(updated)
}

// Delete cancels (soft-deletes) an event — host only.
func (ctrl *Controller) Delete(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)
	id, err := parseEventID(c)
	if err != nil {
		return err
	}
	if err := ctrl.repo.Delete(id, sd.Email); err != nil {
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Invite adds new invitees to an event — host only.
func (ctrl *Controller) Invite(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)
	id, err := parseEventID(c)
	if err != nil {
		return err
	}

	var req inviteRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate all emails exist in the system
	for _, email := range req.Emails {
		if _, err := ctrl.userRepo.Get(email); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "user not found: "+email)
		}
	}

	if err := ctrl.repo.Invite(id, sd.Email, req.Emails); err != nil {
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// RSVP updates the current user's RSVP status for an event.
func (ctrl *Controller) RSVP(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)
	id, err := parseEventID(c)
	if err != nil {
		return err
	}

	var req rsvpRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Status != database.RSVPGoing && req.Status != database.RSVPNotGoing {
		return fiber.NewError(fiber.StatusBadRequest, "status must be 'going' or 'not_going'")
	}

	if err := ctrl.repo.UpdateRSVP(id, sd.Email, req.Status); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
