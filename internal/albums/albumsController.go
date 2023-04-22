package albums

import "github.com/gofiber/fiber/v2"

// todo: use from context instead
const token = ""

func GetAlbums(c *fiber.Ctx) error {
	albums := listAllAlbums(token)

	return c.Status(200).JSON(albums)
}

func CreateAlbum(c *fiber.Ctx) error {
	l := createNewAlbum(token)
	c.Location(l)
	return c.SendStatus(201)
}
