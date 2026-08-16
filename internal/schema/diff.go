package schema

type ChangeType string

const (
	FieldAdded       ChangeType = "added"
	FieldRemoved     ChangeType = "removed"
	FieldTypeChanged ChangeType = "type_changed"
)

type Change struct {
	Path    string
	Type    ChangeType
	OldType FieldType
	NewType FieldType
}

type Diff struct {
	Endpoint string
	Changes  []Change
	Breaking bool
}

func Compare(old, new Schema) []Change {
	var changes []Change

	for path, oldType := range old {
		newType, stillExists := new[path]
		if !stillExists {
			changes = append(changes, Change{
				Path:    path,
				Type:    FieldRemoved,
				OldType: oldType,
			})
			continue
		}
		if newType != oldType {
			changes = append(changes, Change{
				Path:    path,
				Type:    FieldTypeChanged,
				OldType: oldType,
				NewType: newType,
			})
		}
	}

	for path, newType := range new {
		if _, existedBefore := old[path]; !existedBefore {
			changes = append(changes, Change{
				Path:    path,
				Type:    FieldAdded,
				NewType: newType,
			})
		}
	}

	return changes
}

func NewDiff(endpoint string, changes []Change) *Diff {
	if len(changes) == 0 {
		return nil
	}
	d := &Diff{Endpoint: endpoint, Changes: changes}
	for _, c := range changes {
		if c.Type == FieldRemoved || c.Type == FieldTypeChanged {
			d.Breaking = true
			break
		}
	}
	return d
}
