package models

type FieldType string

const (
	FieldText	 FieldType = "text"
	FieldMultiple FieldType = "multiple"
	FieldCheckbox FieldType = "checkbox"
	FieldRating   FieldType = "rating"
)

type FormField struct {
	ID       string    `bson:"id" json:"id"`
	Type     FieldType `bson:"type" json:"type"`
	Label    string    `bson:"label" json:"label"`
	Required bool      `bson:"required" json:"required"`

	Options []string `bson:"options,omitempty" json:"options,omitempty"`

	Max int `bson:"max,omitempty" json:"max,omitempty"`
}

type Form struct {
	ID     string      `bson:"_id" json:"id"`
	Title  string      `bson:"title" json:"title"`
	Fields []FormField `bson:"fields" json:"fields"`
	Status string      `bson:"status" json:"status"`
}

type Response struct {
	ID      string                 `bson:"_id" json:"id"`
	FormID  string                 `bson:"formId" json:"formId"`
	Answers map[string]interface{} `bson:"answers" json:"answers"`
	Created int64                  `bson:"created" json:"created"`
}