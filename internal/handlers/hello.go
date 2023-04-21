package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type SomeStruct struct {
	Hello string `json:"name"`
	Value int    `json:"test"`
}

func HelloWorld(c *fiber.Ctx) error {
	return c.SendString("hello world")
}

func AnotherFunction(c *fiber.Ctx) error {
	//println(c.Locals("session"))
	data := SomeStruct{
		Hello: "world",
		Value: 10,
	}

	return c.JSON(data)
}
