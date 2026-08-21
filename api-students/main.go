package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/students", GetStudents)

	app.Get("/students/:id", GetStudentByID)

	app.Post("/students", CreateStudent)

	app.Put("/students/:id", UpdateStudent)

	app.Patch("/students/:id", PatchStudent)

	app.Delete("/students/:id", DeleteStudent)

	log.Println("Server running on http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
