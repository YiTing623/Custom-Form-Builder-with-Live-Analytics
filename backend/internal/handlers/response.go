package handlers

import (
	"context"
	"fmt"
	"time"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/db"
	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/models"
)

type ResponseHandler struct {
	Store     *db.MongoStore
	Broadcast func(string, []byte)
}

func NewResponseHandler(s *db.MongoStore, broadcaster func(string, []byte)) *ResponseHandler {
	return &ResponseHandler{Store: s, Broadcast: broadcaster}
}

func (h *ResponseHandler) SubmitResponse(c *fiber.Ctx) error {
	formID := c.Params("id")

	var form models.Form
	{
		ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
		defer cancel()
		if err := h.Store.Forms.FindOne(ctx, bson.M{"_id": formID}).Decode(&form); err != nil {
			return fiber.NewError(fiber.StatusNotFound, "form not found")
		}
		if form.Status != "published" {
			return fiber.NewError(fiber.StatusForbidden, "form not published")
		}
	}

	var body models.Response
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	body.ID = uuid.NewString()
	body.FormID = formID
	body.Created = time.Now().Unix()

	if err := validateAnswers(&form, body.Answers); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("validation error: %v", err))
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()
	if _, err := h.Store.Responses.InsertOne(ctx, body); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if h.Broadcast != nil {
		analytics, _ := computeAnalytics(ctx, h.Store, formID, &form)
		msg := fiber.Map{
			"type":      "response:new",
			"formId":    formID,
			"created":   body.Created,
			"analytics": analytics,
		}
		b, _ := json.Marshal(msg)
		h.Broadcast(formID, b)		
	}

	return c.Status(fiber.StatusCreated).JSON(body)
}

func validateAnswers(form *models.Form, ans map[string]interface{}) error {
	for _, f := range form.Fields {
		v, ok := ans[f.ID]

		if f.Required && (!ok || isEmpty(v)) {
			return fmt.Errorf("field '%s' is required", f.ID)
		}
		if !ok {
			continue
		}

		switch f.Type {
		case models.FieldText:
			if _, ok := v.(string); !ok {
				return fmt.Errorf("field '%s' must be string", f.ID)
			}
		case models.FieldMultiple:
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("field '%s' must be string", f.ID)
			}
			if !contains(f.Options, str) {
				return fmt.Errorf("field '%s' must be one of %v", f.ID, f.Options)
			}
		case models.FieldCheckbox:
			arr, ok := toStringSlice(v)
			if !ok {
				return fmt.Errorf("field '%s' must be array of strings", f.ID)
			}
			for _, item := range arr {
				if !contains(f.Options, item) {
					return fmt.Errorf("field '%s' contains invalid option '%s'", f.ID, item)
				}
			}
		case models.FieldRating:
			n, ok := toFloat64(v)
			if !ok {
				return fmt.Errorf("field '%s' must be number", f.ID)
			}
			max := f.Max
			if max <= 0 {
				max = 5
			}
			if n < 1 || n > float64(max) {
				return fmt.Errorf("field '%s' rating must be between 1 and %d", f.ID, max)
			}
		default:
			return fmt.Errorf("unknown field type '%s'", f.Type)
		}
	}
	return nil
}

func isEmpty(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return t == ""
	case []interface{}:
		return len(t) == 0
	case []string:
		return len(t) == 0
	default:
		return false
	}
}

func toStringSlice(v interface{}) ([]string, bool) {
	switch x := v.(type) {
	case []string:
		return x, true
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, e := range x {
			s, ok := e.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	case primitive.A:
		out := make([]string, 0, len(x))
		for _, e := range x {
			s, ok := e.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	default:
		return nil, false
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int, int32, int64:
		return float64(toInt64(n)), true
	default:
		return 0, false
	}
}

func toInt64(v interface{}) int64 {
	switch i := v.(type) {
	case int:
		return int64(i)
	case int32:
		return int64(i)
	case int64:
		return i
	default:
		return 0
	}
}

func contains(opts []string, s string) bool {
	for _, o := range opts {
		if o == s {
			return true
		}
	}
	return false
}
