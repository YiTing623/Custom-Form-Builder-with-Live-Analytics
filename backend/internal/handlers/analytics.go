package handlers

import (
	"context"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/db"
	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/models"
)

type AnalyticsHandler struct {
	Store *db.MongoStore
}

func NewAnalyticsHandler(s *db.MongoStore) *AnalyticsHandler { return &AnalyticsHandler{Store: s} }

func (h *AnalyticsHandler) GetAnalytics(c *fiber.Ctx) error {
	formID := c.Params("id")

	var form models.Form
	{
		ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
		defer cancel()
		if err := h.Store.Forms.FindOne(ctx, bson.M{"_id": formID}).Decode(&form); err != nil {
			return fiber.NewError(fiber.StatusNotFound, "form not found")
		}
	}

	ctx, cancel := context.WithTimeout(c.Context(), 20*time.Second)
	defer cancel()

	out, err := computeAnalytics(ctx, h.Store, formID, &form)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(out)
}

type Analytics struct {
	FormID string                 `json:"formId"`
	Count  int                    `json:"count"`
	Fields map[string]interface{} `json:"fields"`
}

func computeAnalytics(ctx context.Context, store *db.MongoStore, formID string, form *models.Form) (*Analytics, error) {
	cursor, err := store.Responses.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"formId": formID}},
		bson.M{"$project": bson.M{"answers": 1}},
	})
	if err != nil {
		return nil, err
	}
	var rows []bson.M
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	fields := map[string]interface{}{}
	for _, f := range form.Fields {
		switch f.Type {
		case models.FieldMultiple:
			counts := map[string]int{}
			for _, opt := range f.Options {
				counts[opt] = 0
			}
			for _, r := range rows {
				ans := r["answers"].(bson.M)
				if v, ok := ans[f.ID]; ok {
					if s, ok := v.(string); ok {
						if _, exist := counts[s]; exist {
							counts[s]++
						}
					}
				}
			}
			fields[f.ID] = fiber.Map{"type": f.Type, "distribution": counts}

		case models.FieldCheckbox:
			counts := map[string]int{}
			for _, opt := range f.Options {
				counts[opt] = 0
			}
			for _, r := range rows {
				ans := r["answers"].(bson.M)
				if v, ok := ans[f.ID]; ok {
					arr, ok := toStringSlice(v)
					if ok {
						for _, s := range arr {
							if _, exist := counts[s]; exist {
								counts[s]++
							}
						}
					}
				}
			}
			fields[f.ID] = fiber.Map{"type": f.Type, "distribution": counts}

		case models.FieldRating:
			max := f.Max
			if max <= 0 {
				max = 5
			}
			dist := make(map[int]int, max)
			sum := 0.0
			n := 0
			for i := 1; i <= max; i++ {
				dist[i] = 0
			}
			for _, r := range rows {
				ans := r["answers"].(bson.M)
				if v, ok := ans[f.ID]; ok {
					if num, ok := toFloat64(v); ok {
						iv := int(math.Round(num))
						if iv >= 1 && iv <= max {
							dist[iv]++
							sum += num
							n++
						}
					}
				}
			}
			avg := 0.0
			if n > 0 {
				avg = sum / float64(n)
			}
			fields[f.ID] = fiber.Map{"type": f.Type, "distribution": dist, "average": avg}

		case models.FieldText:
			count := 0
			for _, r := range rows {
				ans := r["answers"].(bson.M)
				if v, ok := ans[f.ID]; ok {
					if s, ok := v.(string); ok && s != "" {
						count++
					}
				}
			}
			fields[f.ID] = fiber.Map{"type": f.Type, "nonEmptyCount": count}
		}
	}

	return &Analytics{
		FormID: formID,
		Count:  len(rows),
		Fields: fields,
	}, nil
}
