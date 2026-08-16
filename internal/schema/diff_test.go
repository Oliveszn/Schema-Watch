package schema

import "testing"

func TestCompare_NoChange(t *testing.T) {
	old := Schema{"id": TypeNumber, "email": TypeString}
	new := Schema{"id": TypeNumber, "email": TypeString}

	changes := Compare(old, new)
	if len(changes) != 0 {
		t.Fatalf("expected no changes, got %v", changes)
	}
}

func TestCompare_FieldAdded(t *testing.T) {
	old := Schema{"id": TypeNumber}
	new := Schema{"id": TypeNumber, "email": TypeString}

	changes := Compare(old, new)
	if len(changes) != 1 || changes[0].Type != FieldAdded || changes[0].Path != "email" {
		t.Fatalf("expected one FieldAdded change for 'email', got %v", changes)
	}
}

func TestCompare_FieldRemoved(t *testing.T) {
	old := Schema{"id": TypeNumber, "email": TypeString}
	new := Schema{"id": TypeNumber}

	changes := Compare(old, new)
	if len(changes) != 1 || changes[0].Type != FieldRemoved || changes[0].Path != "email" {
		t.Fatalf("expected one FieldRemoved change for 'email', got %v", changes)
	}
}

func TestCompare_FieldTypeChanged(t *testing.T) {
	old := Schema{"id": TypeNumber}
	new := Schema{"id": TypeString}

	changes := Compare(old, new)
	if len(changes) != 1 || changes[0].Type != FieldTypeChanged {
		t.Fatalf("expected one FieldTypeChanged change, got %v", changes)
	}
	if changes[0].OldType != TypeNumber || changes[0].NewType != TypeString {
		t.Fatalf("unexpected old/new types: %+v", changes[0])
	}
}

func TestNewDiff_BreakingOnRemoval(t *testing.T) {
	changes := []Change{{Path: "email", Type: FieldRemoved, OldType: TypeString}}
	d := NewDiff("GET /users", changes)
	if d == nil || !d.Breaking {
		t.Fatalf("expected breaking diff, got %+v", d)
	}
}

func TestNewDiff_NotBreakingOnAddition(t *testing.T) {
	changes := []Change{{Path: "nickname", Type: FieldAdded, NewType: TypeString}}
	d := NewDiff("GET /users", changes)
	if d == nil || d.Breaking {
		t.Fatalf("expected non-breaking diff, got %+v", d)
	}
}

func TestNewDiff_NilWhenNoChanges(t *testing.T) {
	d := NewDiff("GET /users", nil)
	if d != nil {
		t.Fatalf("expected nil diff for no changes, got %+v", d)
	}
}
