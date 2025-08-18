package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	if body.Answers == nil {
		body.Answers = map[string]interface{}{}
	}
	body.ID = uuid.NewString()
	body.FormID = formID
	body.Created = time.Now().Unix()

	visible := computeVisibility(&form, body.Answers)

	if err := validateAnswers(&form, body.Answers, visible); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("validation error: %v", err))
	}
	for _, f := range form.Fields {
		if !visible[f.ID] {
			delete(body.Answers, f.ID)
		}
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


func computeVisibility(form *models.Form, answers map[string]interface{}) map[string]bool {
	vis := make(map[string]bool, len(form.Fields))
	for _, f := range form.Fields {
		if f.ShowIf == nil {
			vis[f.ID] = true
			continue
		}
		vis[f.ID] = evalCondition(f.ShowIf, answers)
	}
	return vis
}

func evalCondition(cond *models.ShowIf, answers map[string]interface{}) bool {
	if cond == nil || cond.FieldID == "" {
		return true
	}
	v, ok := answers[cond.FieldID]
	if !ok {
		return false
	}
	switch cond.Operator {
	case models.OpEq:
		return equalVal(v, cond.Value)
	case models.OpNe:
		return !equalVal(v, cond.Value)
	case models.OpIncludes:
		return arrIncludes(v, cond.Value)
	case models.OpGt, models.OpGte, models.OpLt, models.OpLte:
		return numericCompare(v, cond.Value, string(cond.Operator))
	default:
		return false
	}
}

func equalVal(a, b interface{}) bool {
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		return strings.TrimSpace(as) == strings.TrimSpace(bs)
	}
	af, okA := toFloat64(a)
	bf, okB := toFloat64(b)
	if okA && okB {
		return af == bf
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func arrIncludes(arr interface{}, needle interface{}) bool {
	switch x := arr.(type) {
	case []string:
		ns := fmt.Sprintf("%v", needle)
		for _, e := range x {
			if e == ns {
				return true
			}
		}
		return false
	case []interface{}:
		ns := fmt.Sprintf("%v", needle)
		for _, e := range x {
			if fmt.Sprintf("%v", e) == ns {
				return true
			}
		}
		return false
	default:
		return equalVal(arr, needle)
	}
}

func numericCompare(a, b interface{}, op string) bool {
	af, ok1 := toFloat64(a)
	bf, ok2 := toFloat64(b)
	if !ok1 || !ok2 {
		return false
	}
	switch op {
	case "gt":
		return af > bf
	case "gte":
		return af >= bf
	case "lt":
		return af < bf
	case "lte":
		return af <= bf
	default:
		return false
	}
}


func validateAnswers(form *models.Form, ans map[string]interface{}, visible map[string]bool) error {
	for _, f := range form.Fields {
		vis := visible[f.ID]
		v, ok := ans[f.ID]

		if vis && f.Required && (!ok || isEmpty(v)) {
			return fmt.Errorf("field '%s' is required", f.ID)
		}
		if !ok {
			continue
		}
		if !vis {
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
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
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
