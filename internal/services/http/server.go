package http

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"log"
)

type Server struct {
	Addr string
	app  fiber.App
}

func (s *Server) Run(ctx context.Context) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Token Handler",
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	go func() {
		select {
		case <-ctx.Done():
			log.Println("Token local server shutting down...")
			log.Fatal(app.Shutdown())
		}
	}()

	go func() {
		log.Println("Token local server starting...")
		log.Fatal(app.Listen(s.Addr))
	}()
}
