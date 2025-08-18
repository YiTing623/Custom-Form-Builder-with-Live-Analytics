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
	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/middleware"
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
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,OPTIONS",
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
	exportH := handlers.NewExportHandler(store)

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("dev_change_me")
	}
	authH := handlers.NewAuthHandler(store, jwtSecret)

	api := app.Group("/api")

	api.Post("/auth/register", authH.Register)
	api.Post("/auth/login", authH.Login)

	public := api.Group("", middleware.AuthOptional(jwtSecret))
	public.Get("/forms/:id", formH.GetForm)
	public.Get("/forms/:id/analytics", analyticsH.GetAnalytics)
	public.Get("/forms/:id/export", exportH.ExportResponses)
	public.Post("/forms/:id/response", respH.SubmitResponse)

	priv := api.Group("", middleware.AuthRequired(jwtSecret))
	priv.Get("/me", authH.Me)
	priv.Get("/my/forms", formH.ListMyForms)
	priv.Post("/forms", formH.CreateForm)
	priv.Put("/forms/:id", formH.UpdateForm)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(app.Listen(":" + port))
}
