package contacts

import (
	"eventapi/internal/database"
	"eventapi/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type addContactRequest struct {
	Email string `json:"email"`
}

// ─── Controller ───────────────────────────────────────────────────────────────

type Controller struct {
	repo     database.ContactRepository
	userRepo database.UserRepository
}

func NewController(repo database.ContactRepository, userRepo database.UserRepository) *Controller {
	return &Controller{repo: repo, userRepo: userRepo}
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// List returns all contacts for the current user.
func (ctrl *Controller) List(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)
	contacts, err := ctrl.repo.List(sd.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(contacts)
}

// Add adds a user to the current user's contacts by email.
func (ctrl *Controller) Add(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)

	var req addContactRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email is required")
	}

	// Verify the contact user exists
	if _, err := ctrl.userRepo.Get(req.Email); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	contact, err := ctrl.repo.Add(sd.Email, req.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(contact)
}

// Remove deletes a contact by email.
func (ctrl *Controller) Remove(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)
	contactEmail := c.Params("email")
	if contactEmail == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email parameter required")
	}
	if err := ctrl.repo.Remove(sd.Email, contactEmail); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
