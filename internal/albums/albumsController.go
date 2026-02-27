package albums

import (
	"eventapi/internal/database"
	"eventapi/internal/middleware"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ─── Controller ───────────────────────────────────────────────────────────────

type Controller struct {
	eventRepo database.EventRepository
}

func NewController(eventRepo database.EventRepository) *Controller {
	return &Controller{eventRepo: eventRepo}
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// UploadPhotos handles multipart photo uploads for an event album.
// Phase 1: accepts the request and stores placeholder EventPhoto records.
// Phase 6: will upload bytes to Google Photos using the user's access token.
func (ctrl *Controller) UploadPhotos(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)

	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid event ID")
	}

	// Verify event exists
	event, err := ctrl.eventRepo.GetByID(eventID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "event not found")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "expected multipart form")
	}

	files := form.File["photos"]
	if len(files) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "no photos provided")
	}

	var uploaded []database.EventPhoto
	for _, fileHeader := range files {
		// TODO (Phase 6): read file bytes, call UploadPhoto + AddPhotoToAlbum
		// using sd.AccessToken and event.AlbumID
		_ = event // will be used in Phase 6 for album upload

		photo := database.EventPhoto{
			EventID:       eventID,
			UploaderEmail: sd.Email,
			GooglePhotoID: fmt.Sprintf("pending-%s", fileHeader.Filename),
		}
		saved, err := ctrl.eventRepo.AddPhoto(photo)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		uploaded = append(uploaded, *saved)
	}

	return c.Status(fiber.StatusCreated).JSON(uploaded)
}

// DeletePhoto removes a photo the current user uploaded.
func (ctrl *Controller) DeletePhoto(c *fiber.Ctx) error {
	sd := middleware.GetSession(c)

	photoID, err := uuid.Parse(c.Params("photoId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid photo ID")
	}

	if err := ctrl.eventRepo.DeletePhoto(photoID, sd.Email); err != nil {
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
