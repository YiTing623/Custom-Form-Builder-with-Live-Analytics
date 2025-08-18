package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"

	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/db"
	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExportHandler struct {
	Store *db.MongoStore
}

func NewExportHandler(s *db.MongoStore) *ExportHandler {
	return &ExportHandler{Store: s}
}

func (h *ExportHandler) ExportResponses(c *fiber.Ctx) error {
	formID := c.Params("id")
	format := strings.ToLower(c.Query("format", "csv"))

	var form models.Form
	if err := h.Store.Forms.FindOne(c.Context(), bson.M{"_id": formID}).Decode(&form); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "form not found")
	}

	cur, err := h.Store.Responses.Find(
		c.Context(),
		bson.M{"formId": formID},
		&options.FindOptions{Sort: bson.M{"created": 1}},
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	defer cur.Close(c.Context())

	var resps []models.Response
	for cur.Next(c.Context()) {
		var r models.Response
		if err := cur.Decode(&r); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "decode error")
		}
		resps = append(resps, r)
	}

	filename := sanitizeFilename(fmt.Sprintf("responses-%s-%s", formID, time.Now().Format("20060102-150405")))
	switch format {
	case "csv":
		data, err := h.renderCSV(&form, resps)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "csv render error")
		}
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, url.PathEscape(filename)))
		return c.Send(data)

	case "pdf":
		data, err := h.renderPDF(&form, resps)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "pdf render error")
		}
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, url.PathEscape(filename)))
		return c.Send(data)

	default:
		return fiber.NewError(fiber.StatusBadRequest, "supported formats: csv, pdf")
	}
}

func (h *ExportHandler) renderCSV(form *models.Form, resps []models.Response) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{"created"}
	for _, f := range form.Fields {
		header = append(header, f.Label)
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	// rows
	for _, r := range resps {
		row := []string{time.Unix(r.Created, 0).Format(time.RFC3339)}
		for _, f := range form.Fields {
			val := renderAnswerCSV(r.Answers[f.ID])
			row = append(row, val)
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderAnswerCSV(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", x), "0"), ".")
	case int, int64:
		return fmt.Sprintf("%v", x)
	case []string:
		return strings.Join(x, "; ")
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, e := range x {
			out = append(out, fmt.Sprintf("%v", e))
		}
		return strings.Join(out, "; ")
	default:
		return fmt.Sprintf("%v", x)
	}
}

func (h *ExportHandler) renderPDF(form *models.Form, resps []models.Response) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Form Responses", false)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(0, 10, form.Title)
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 11)
	meta := fmt.Sprintf("Form ID: %s   Responses: %d   Generated: %s",
		form.ID, len(resps), time.Now().Format("2006-01-02 15:04:05"))
	pdf.Cell(0, 8, meta)
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 11)
	cols := []string{"Created"}
	for _, f := range form.Fields {
		cols = append(cols, f.Label)
	}

	colWidths := autoColumnWidths(pdf, cols, resps, form, 190)
	for i, htxt := range cols {
		pdf.CellFormat(colWidths[i], 8, htxt, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 10)
	for _, r := range resps {
		cells := []string{time.Unix(r.Created, 0).Format("2006-01-02 15:04")}
		for _, f := range form.Fields {
			cells = append(cells, renderAnswerPDF(r.Answers[f.ID]))
		}
		maxLines := 1
		lineHeights := make([]int, len(cells))
		lines := make([][]string, len(cells))
		for i, cell := range cells {
			wrap := pdf.SplitLines([]byte(cell), colWidths[i])
			lines[i] = make([]string, len(wrap))
			for j := range wrap {
				lines[i][j] = string(wrap[j])
			}
			if len(wrap) > maxLines {
				maxLines = len(wrap)
			}
		}
		h := 6.0
		for line := 0; line < maxLines; line++ {
			for i := range cells {
				text := ""
				if line < len(lines[i]) {
					text = lines[i][line]
				}
				border := "LR"
				if line == 0 {
					border = "LTR"
				}
				if line == maxLines-1 {
					border = "LBR"
				}
				pdf.CellFormat(colWidths[i], h, text, border, 0, "L", false, 0, "")
				lineHeights[i]++
			}
			pdf.Ln(h)
		}
	}

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func renderAnswerPDF(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", x), "0"), ".")
	case int, int64:
		return fmt.Sprintf("%v", x)
	case []string:
		return strings.Join(x, ", ")
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, e := range x {
			out = append(out, fmt.Sprintf("%v", e))
		}
		return strings.Join(out, ", ")
	default:
		return fmt.Sprintf("%v", x)
	}
}

func autoColumnWidths(pdf *gofpdf.Fpdf, header []string, resps []models.Response, form *models.Form, maxWidth float64) []float64 {
	n := len(header)
	widths := make([]float64, n)
	min := 20.0
	for i := range widths {
		widths[i] = min
	}

	measure := func(s string) float64 {
		return pdf.GetStringWidth(s) + 6
	}

	pdf.SetFont("Helvetica", "B", 11)
	for i, h := range header {
		if w := measure(h); w > widths[i] {
			widths[i] = w
		}
	}

	pdf.SetFont("Helvetica", "", 10)
	limit := 50
	if len(resps) < limit {
		limit = len(resps)
	}
	for idx := 0; idx < limit; idx++ {
		r := resps[idx]
		cells := []string{time.Unix(r.Created, 0).Format("2006-01-02 15:04")}
		for _, f := range form.Fields {
			cells = append(cells, renderAnswerPDF(r.Answers[f.ID]))
		}
		for i, txt := range cells {
			if w := measure(txt); w > widths[i] {
				widths[i] = w
			}
		}
	}

	total := 0.0
	for _, w := range widths {
		total += w
	}
	if total > maxWidth {
		scale := maxWidth / total
		for i := range widths {
			widths[i] = widths[i] * scale
			if widths[i] < 16 {
				widths[i] = 16
			}
		}
	}

	return widths
}

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	return s
}
