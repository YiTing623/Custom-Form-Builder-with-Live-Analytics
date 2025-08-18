package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/db"
	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/handlers"
	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/ws"
)

func main() {
	_ = godotenv.Load()

	store, err := db.NewMongoStore()
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer store.Client.Disconnect(nil)

	app := fiber.New(fiber.Config{
		ReadTimeout:  0,
		WriteTimeout: 0,
		AppName:      "FormBuilder API",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/api/health", func(c *fiber.Ctx) error { return c.SendString("ok") })

	hub := ws.NewHub()

	app.Get("/api/sse/:formId", func(c *fiber.Ctx) error {
		formID := c.Params("formId")

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("X-Accel-Buffering", "no")

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			sub := hub.Subscribe(formID)
			defer hub.Unsubscribe(formID, sub)

			hello, _ := json.Marshal(fiber.Map{"type": "hello", "ts": time.Now().Unix()})
			w.WriteString("event: message\n")
			w.WriteString("data: ")
			w.Write(hello)
			w.WriteString("\n\n")
			_ = w.Flush()

			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case msg, ok := <-sub:
					if !ok {
						return
					}
					w.WriteString("event: message\n")
					w.WriteString("data: ")
					w.Write(msg)
					w.WriteString("\n\n")
					if err := w.Flush(); err != nil {
						return
					}
				case <-ticker.C:
					w.WriteString(": ping\n\n")
					if err := w.Flush(); err != nil {
						return
					}
				}
			}
		})

		return nil
	})

	broadcast := func(formID string, payload []byte) {
		hub.Broadcast(formID, payload)
	}

	formH := handlers.NewFormHandler(store)
	respH := handlers.NewResponseHandler(store, broadcast)
	analyticsH := handlers.NewAnalyticsHandler(store)

	api := app.Group("/api")
	api.Post("/forms", formH.CreateForm)
	api.Get("/forms/:id", formH.GetForm)
	api.Put("/forms/:id", formH.UpdateForm)

	api.Post("/forms/:id/response", respH.SubmitResponse)
	api.Get("/forms/:id/analytics", analyticsH.GetAnalytics)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(app.Listen(":" + port))
}
