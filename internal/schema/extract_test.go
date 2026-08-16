package schema

import "testing"

func TestExtract_FlatObject(t *testing.T) {
	data := []byte(`{"id": 1, "email": "a@b.com", "active": true}`)

	got, err := Extract(data)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	want := Schema{
		"id":     TypeNumber,
		"email":  TypeString,
		"active": TypeBool,
	}
	assertSchemaEqual(t, got, want)
}

func TestExtract_NestedObject(t *testing.T) {
	data := []byte(`{"user": {"id": 1, "address": {"city": "Lagos"}}}`)

	got, err := Extract(data)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	want := Schema{
		"user.id":           TypeNumber,
		"user.address.city": TypeString,
	}
	assertSchemaEqual(t, got, want)
}

func TestExtract_ArrayOfObjects(t *testing.T) {
	data := []byte(`{"items": [{"id": 1, "name": "a"}, {"id": 2, "name": "b"}]}`)

	got, err := Extract(data)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	want := Schema{
		"items[].id":   TypeNumber,
		"items[].name": TypeString,
	}
	assertSchemaEqual(t, got, want)
}

func TestExtract_ArrayOfPrimitives(t *testing.T) {
	data := []byte(`{"tags": ["go", "gin"]}`)

	got, err := Extract(data)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	want := Schema{"tags[]": TypeString}
	assertSchemaEqual(t, got, want)
}

func TestExtract_EmptyArrayAndObject(t *testing.T) {
	data := []byte(`{"items": [], "meta": {}}`)

	got, err := Extract(data)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	want := Schema{
		"items[]": TypeUnknown,
		"meta":    TypeObject,
	}
	assertSchemaEqual(t, got, want)
}

func TestExtract_NullField(t *testing.T) {
	data := []byte(`{"deleted_at": null}`)

	got, err := Extract(data)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	want := Schema{"deleted_at": TypeNull}
	assertSchemaEqual(t, got, want)
}

func TestExtract_InvalidJSON(t *testing.T) {
	_, err := Extract([]byte(`{not valid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func assertSchemaEqual(t *testing.T, got, want Schema) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("schema length mismatch: got %d fields %v, want %d fields %v", len(got), got, len(want), want)
	}
	for path, wantType := range want {
		gotType, ok := got[path]
		if !ok {
			t.Errorf("missing path %q in extracted schema", path)
			continue
		}
		if gotType != wantType {
			t.Errorf("path %q: got type %q, want %q", path, gotType, wantType)
		}
	}
}
