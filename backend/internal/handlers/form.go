package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/db"
	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/models"
)

type FormHandler struct {
	Store *db.MongoStore
}

func NewFormHandler(s *db.MongoStore) *FormHandler { return &FormHandler{Store: s} }

func (h *FormHandler) CreateForm(c *fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)
	if userID == "" {
		return fiber.ErrUnauthorized
	}

	var body models.Form
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if body.ID == "" {
		body.ID = uuid.NewString()
	}
	if body.Title = strings.TrimSpace(body.Title); body.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	if body.Status == "" {
		body.Status = "draft"
	}
	for i := range body.Fields {
		if err := validateField(&body.Fields[i]); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("fields[%d]: %v", i, err))
		}
	}

	body.OwnerID = userID

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()
	if _, err := h.Store.Forms.InsertOne(ctx, body); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(body)
}

func (h *FormHandler) GetForm(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	var form models.Form
	if err := h.Store.Forms.FindOne(ctx, bson.M{"_id": id}).Decode(&form); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "form not found")
	}

	if form.Status != "published" && form.OwnerID != userID {
		return fiber.ErrForbidden
	}
	return c.JSON(form)
}

func (h *FormHandler) UpdateForm(c *fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)
	if userID == "" {
		return fiber.ErrUnauthorized
	}
	id := c.Params("id")

	{
		ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
		defer cancel()

		var exist models.Form
		if err := h.Store.Forms.FindOne(ctx, bson.M{"_id": id}).Decode(&exist); err != nil {
			return fiber.NewError(fiber.StatusNotFound, "form not found")
		}
		if exist.OwnerID != userID {
			return fiber.ErrForbidden
		}
	}

	var body models.Form
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	body.ID = id
	if body.Title = strings.TrimSpace(body.Title); body.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	if body.Status == "" {
		body.Status = "draft"
	}
	for i := range body.Fields {
		if err := validateField(&body.Fields[i]); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("fields[%d]: %v", i, err))
		}
	}

	update := bson.M{
		"title":   body.Title,
		"fields":  body.Fields,
		"status":  body.Status,
		"ownerId": userID,
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()
	_, err := h.Store.Forms.UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	body.OwnerID = userID
	return c.JSON(body)
}

func (h *FormHandler) ListMyForms(c *fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)
	if userID == "" {
		return fiber.ErrUnauthorized
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	cur, err := h.Store.Forms.Find(ctx, bson.M{"ownerId": userID}, &options.FindOptions{
		Sort: bson.M{"_id": -1},
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer cur.Close(ctx)

	var out []models.Form
	for cur.Next(ctx) {
		var f models.Form
		if err := cur.Decode(&f); err == nil {
			out = append(out, f)
		}
	}
	return c.JSON(out)
}

func validateField(f *models.FormField) error {
	if f.ID = strings.TrimSpace(f.ID); f.ID == "" {
		return fmt.Errorf("id is required")
	}
	if f.Label = strings.TrimSpace(f.Label); f.Label == "" {
		return fmt.Errorf("label is required")
	}
	switch f.Type {
	case models.FieldText:
	case models.FieldMultiple:
		if len(f.Options) == 0 {
			return fmt.Errorf("multiple requires non-empty options")
		}
	case models.FieldCheckbox:
		if len(f.Options) == 0 {
			return fmt.Errorf("checkbox requires non-empty options")
		}
	case models.FieldRating:
		if f.Max <= 0 {
			f.Max = 5
		}
	default:
		return fmt.Errorf("unknown type: %s", f.Type)
	}
	return nil
}
