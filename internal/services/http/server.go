package http

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"log"
)

type Server struct {
	addr     string
	codeChan chan string
	app      fiber.App
}

func NewServer(addr string, codeChan chan string) Server {
	return Server{
		addr:     addr,
		codeChan: codeChan,
	}
}

func (s *Server) Run(ctx context.Context) {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Token Handler",
		Views:                 engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		qCode := c.Query("code")
		if qCode != "" {
			if len(s.codeChan) == 0 {
				s.codeChan <- qCode
				return c.Render("index", fiber.Map{
					"Title":   "Retrieving code",
					"Message": "Done!",
				})
			}
		}

		return c.SendString("Nothing here...")
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
		log.Fatal(app.Listen(s.addr))
	}()
}
