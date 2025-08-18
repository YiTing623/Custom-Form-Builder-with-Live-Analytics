package models

type FieldType string

const (
	FieldText	 FieldType = "text"
	FieldMultiple FieldType = "multiple"
	FieldCheckbox FieldType = "checkbox"
	FieldRating   FieldType = "rating"
)

type ConditionOperator string

const (
	OpEq ConditionOperator = "eq"
	OpNe ConditionOperator = "ne"
	OpIncludes ConditionOperator = "includes"
	OpGt ConditionOperator = "gt"
	OpLt ConditionOperator = "lt"
	OpGte ConditionOperator = "gte"
	OpLte ConditionOperator = "lte"

)

type ShowIf struct {
	FieldID string `bson:"fieldId" json:"fieldId"`
	Operator ConditionOperator `bson:"op" json:"op"`
	Value interface{} `bson:"value" json:"value"`
}

type FormField struct {
	ID       string    `bson:"id" json:"id"`
	Type     FieldType `bson:"type" json:"type"`
	Label    string    `bson:"label" json:"label"`
	Required bool      `bson:"required" json:"required"`

	Options []string `bson:"options,omitempty" json:"options,omitempty"`

	Max int `bson:"max,omitempty" json:"max,omitempty"`
	ShowIf *ShowIf `bson:"showIf,omitempty"  json:"showIf,omitempty"`
}

type Form struct {
	ID     string      `bson:"_id" json:"id"`
	Title  string      `bson:"title" json:"title"`
	Fields []FormField `bson:"fields" json:"fields"`
	Status string      `bson:"status" json:"status"`
	OwnerID string      `bson:"ownerId" json:"ownerId"`
}

type Response struct {
	ID      string                 `bson:"_id" json:"id"`
	FormID  string                 `bson:"formId" json:"formId"`
	UserID  string                 `bson:"userId,omitempty" json:"userId,omitempty"`
	Answers map[string]interface{} `bson:"answers" json:"answers"`
	Created int64                  `bson:"created" json:"created"`
}

type User struct {
	ID           string `bson:"_id"         json:"id"`
	Email        string `bson:"email"       json:"email"`
	Name         string `bson:"name"        json:"name"`
	PasswordHash []byte `bson:"passwordHash" json:"-"`
	Created      int64  `bson:"created"     json:"created"`
}