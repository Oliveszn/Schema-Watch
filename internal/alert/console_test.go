package alert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Oliveszn/Schema-Watch/internal/schema"
)

func TestConsole_Alert_NilDiffIsNoop(t *testing.T) {
	var buf bytes.Buffer
	c := NewConsoleWithWriter(&buf, false)

	c.Alert(nil)

	if buf.Len() != 0 {
		t.Fatalf("expected no output for nil diff, got %q", buf.String())
	}
}

func TestConsole_Alert_BreakingChangeLabel(t *testing.T) {
	var buf bytes.Buffer
	c := NewConsoleWithWriter(&buf, false)

	diff := &schema.Diff{
		Endpoint: "GET /users/1",
		Breaking: true,
		Changes: []schema.Change{
			{Path: "id", Type: schema.FieldTypeChanged, OldType: schema.TypeNumber, NewType: schema.TypeString},
		},
	}
	c.Alert(diff)

	out := buf.String()
	if !strings.Contains(out, "BREAKING CHANGE on GET /users/1") {
		t.Fatalf("expected BREAKING CHANGE header, got %q", out)
	}
	if !strings.Contains(out, "~ id changed type: number -> string") {
		t.Fatalf("expected type-changed line, got %q", out)
	}
}

func TestConsole_Alert_NonBreakingChangeLabel(t *testing.T) {
	var buf bytes.Buffer
	c := NewConsoleWithWriter(&buf, false)

	diff := &schema.Diff{
		Endpoint: "GET /users/1",
		Breaking: false,
		Changes: []schema.Change{
			{Path: "nickname", Type: schema.FieldAdded, NewType: schema.TypeString},
		},
	}
	c.Alert(diff)

	out := buf.String()
	if strings.Contains(out, "BREAKING CHANGE") {
		t.Fatalf("expected non-breaking label, got %q", out)
	}
	if !strings.Contains(out, "CHANGE on GET /users/1") {
		t.Fatalf("expected CHANGE header, got %q", out)
	}
	if !strings.Contains(out, "+ nickname added (string)") {
		t.Fatalf("expected field-added line, got %q", out)
	}
}

func TestConsole_Alert_AllThreeChangeTypes(t *testing.T) {
	var buf bytes.Buffer
	c := NewConsoleWithWriter(&buf, false)

	diff := &schema.Diff{
		Endpoint: "GET /orders/5",
		Breaking: true,
		Changes: []schema.Change{
			{Path: "total", Type: schema.FieldTypeChanged, OldType: schema.TypeNumber, NewType: schema.TypeString},
			{Path: "coupon_code", Type: schema.FieldRemoved, OldType: schema.TypeString},
			{Path: "currency", Type: schema.FieldAdded, NewType: schema.TypeString},
		},
	}
	c.Alert(diff)

	out := buf.String()
	for _, want := range []string{
		"~ total changed type: number -> string",
		"- coupon_code removed (was string)",
		"+ currency added (string)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
}

func TestConsole_Alert_NoColorWhenDisabled(t *testing.T) {
	var buf bytes.Buffer
	c := NewConsoleWithWriter(&buf, false)

	c.Alert(&schema.Diff{
		Endpoint: "GET /x",
		Breaking: true,
		Changes:  []schema.Change{{Path: "a", Type: schema.FieldAdded, NewType: schema.TypeBool}},
	})

	if strings.Contains(buf.String(), "\033[") {
		t.Fatalf("expected no ANSI codes when color disabled, got %q", buf.String())
	}
}

func TestConsole_Alert_ColorWhenEnabled(t *testing.T) {
	var buf bytes.Buffer
	c := NewConsoleWithWriter(&buf, true)

	c.Alert(&schema.Diff{
		Endpoint: "GET /x",
		Breaking: true,
		Changes:  []schema.Change{{Path: "a", Type: schema.FieldAdded, NewType: schema.TypeBool}},
	})

	if !strings.Contains(buf.String(), "\033[") {
		t.Fatalf("expected ANSI codes when color enabled, got %q", buf.String())
	}
}
