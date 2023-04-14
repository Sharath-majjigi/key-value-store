package main

import (
	_"fmt"
	"strings"
	"sharath/request"

	"github.com/gofiber/fiber/v2"
)

func main() {
	db := &request.Database{
		M: make(map[string]*request.KeyValue),
	}

	app := fiber.New()

	app.Post("/", func(c *fiber.Ctx) error {
		var data map[string]string

		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid input data",
			})
		}

		command := data["command"]
		parts := strings.Fields(command)

		if len(parts) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid command",
			})
		}

		switch strings.ToUpper(parts[0]) {
		case "SET":
			return request.HandleSetCommand(c, db, parts[1:])
		case "GET":
			return request.HandleGetCommand(c, db, parts[1:])
		case "QPUSH":
			return request.HandleQPushCommand(c, db, parts[1:])
		case "QPOP":
			return request.HandleQPopCommand(c,db,parts[1:])
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid command",
			})
		}
	})

	app.Listen(":3000")
}
