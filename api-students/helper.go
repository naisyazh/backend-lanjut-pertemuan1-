package main

import "github.com/gofiber/fiber/v2"

func SuccessResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func ErrorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

func isNIMExists(nim string, excludeID int) bool {
	for _, student := range students {
		if student.NIM == nim && student.ID != excludeID {
			return true
		}
	}
	return false
}

func findStudentByID(id int) *Student {
	for i := range students {
		if students[i].ID == id {
			return &students[i]
		}
	}
	return nil
}
