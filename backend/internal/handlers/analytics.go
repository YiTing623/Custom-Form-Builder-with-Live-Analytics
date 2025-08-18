package handlers

import (
	"context"
	"math"
	"sort"
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


type Trends struct {
	AvgRating   float64                `json:"avgRating,omitempty"`
	MostCommon  map[string]interface{} `json:"mostCommon,omitempty"`
	Skipped     map[string]int         `json:"skipped,omitempty"`
	MostSkipped []fiber.Map            `json:"mostSkipped,omitempty"` 
}

type Analytics struct {
	FormID string                 `json:"formId"`
	Count  int                    `json:"count"`
	Fields map[string]interface{} `json:"fields"`
	Trends *Trends                `json:"trends,omitempty"`
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

	total := len(rows)
	fields := map[string]interface{}{}

	skipped := make(map[string]int)
	mostCommon := make(map[string]interface{})
	var globalRatingSum float64
	var globalRatingCount int

	labelOf := func(id string) string {
		for _, f := range form.Fields {
			if f.ID == id {
				return f.Label
			}
		}
		return id
	}

	for _, f := range form.Fields {
		switch f.Type {
		case models.FieldMultiple:
			counts := map[string]int{}
			for _, opt := range f.Options {
				counts[opt] = 0
			}
			seen := 0
			for _, r := range rows {
				ans := r["answers"].(bson.M)
				if v, ok := ans[f.ID]; ok {
					if s, ok := v.(string); ok && s != "" {
						if _, exist := counts[s]; exist {
							counts[s]++
						}
						seen++
					}
				}
			}
			fields[f.ID] = fiber.Map{"type": f.Type, "distribution": counts}
			skipped[f.ID] = total - seen

			topOpt := ""
			topCnt := -1
			for k, v := range counts {
				if v > topCnt {
					topCnt = v
					topOpt = k
				}
			}
			if topCnt >= 0 {
				mostCommon[f.ID] = topOpt
			}

		case models.FieldCheckbox:
			counts := map[string]int{}
			for _, opt := range f.Options {
				counts[opt] = 0
			}
			seen := 0
			for _, r := range rows {
				ans := r["answers"].(bson.M)
				if v, ok := ans[f.ID]; ok {
					arr, ok := toStringSlice(v)
					if ok && len(arr) > 0 {
						for _, s := range arr {
							if _, exist := counts[s]; exist {
								counts[s]++
							}
						}
						seen++
					}
				}
			}
			fields[f.ID] = fiber.Map{"type": f.Type, "distribution": counts}
			skipped[f.ID] = total - seen

			topCnt := -1
			topOptions := []string{}
			for k, v := range counts {
				if v > topCnt {
					topCnt = v
					topOptions = []string{k}
				} else if v == topCnt {
					topOptions = append(topOptions, k)
				}
			}
			if topCnt >= 0 {
				mostCommon[f.ID] = topOptions
			}

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
			skipped[f.ID] = total - n

			globalRatingSum += sum
			globalRatingCount += n

		case models.FieldText:
			nonEmpty := 0
			for _, r := range rows {
				ans := r["answers"].(bson.M)
				if v, ok := ans[f.ID]; ok {
					if s, ok := v.(string); ok && s != "" {
						nonEmpty++
					}
				}
			}
			fields[f.ID] = fiber.Map{"type": f.Type, "nonEmptyCount": nonEmpty}
			skipped[f.ID] = total - nonEmpty
		}
	}

	mostSkipped := make([]fiber.Map, 0, len(form.Fields))
	for _, f := range form.Fields {
		mostSkipped = append(mostSkipped, fiber.Map{
			"id":      f.ID,
			"label":   labelOf(f.ID),
			"skipped": skipped[f.ID],
			"total":   total,
		})
	}
	sort.Slice(mostSkipped, func(i, j int) bool {
		return mostSkipped[i]["skipped"].(int) > mostSkipped[j]["skipped"].(int)
	})
	if len(mostSkipped) > 3 {
		mostSkipped = mostSkipped[:3]
	}

	var avgRating float64
	if globalRatingCount > 0 {
		avgRating = globalRatingSum / float64(globalRatingCount)
	}

	trends := &Trends{
		AvgRating:   avgRating,
		MostCommon:  mostCommon,
		Skipped:     skipped,
		MostSkipped: mostSkipped,
	}

	return &Analytics{
		FormID: formID,
		Count:  total,
		Fields: fields,
		Trends: trends,
	}, nil
}
