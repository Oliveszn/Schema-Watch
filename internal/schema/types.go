package schema

import "time"

type FieldType string

const (
	TypeString  FieldType = "string"
	TypeNumber  FieldType = "number"
	TypeBool    FieldType = "bool"
	TypeNull    FieldType = "null"
	TypeObject  FieldType = "object"
	TypeArray   FieldType = "array"
	TypeUnknown FieldType = "unknown"
)

type Schema map[string]FieldType

type Snapshot struct {
	Endpoint   string
	Schema     Schema
	CapyuredAt time.Time
}
