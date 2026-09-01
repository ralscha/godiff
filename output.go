package godiff

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"strconv"
	"strings"
)

var (
	_ json.Marshaler   = (*DiffResult)(nil)
	_ json.MarshalerTo = (*DiffResult)(nil)
)

// String returns a human-readable representation of the diff result
func (dr *DiffResult) String() string {
	if dr == nil || len(dr.Diffs) == 0 {
		return "No differences found"
	}

	var sb strings.Builder
	sb.Grow(30 + len(dr.Diffs)*90)

	sb.WriteString("Found ")
	sb.WriteString(strconv.Itoa(len(dr.Diffs)))
	sb.WriteString(" differences:\n")

	for _, diff := range dr.Diffs {
		switch d := diff.(type) {
		case *MapDiff:
			sb.WriteString(string(d.ChangeType))
			sb.WriteString(" ")
			sb.WriteString(d.Path)
			sb.WriteString(": ")
			switch d.ChangeType {
			case ChangeTypeAdded:
				fmt.Fprint(&sb, d.Right)
			case ChangeTypeRemoved:
				fmt.Fprint(&sb, d.Left)
			default:
				fmt.Fprint(&sb, d.Left)
				sb.WriteString(" -> ")
				fmt.Fprint(&sb, d.Right)
			}
			sb.WriteString("\n")
		case *SliceDiff:
			sb.WriteString(string(d.ChangeType))
			sb.WriteString(" ")
			sb.WriteString(d.Path)
			sb.WriteString("[")
			sb.WriteString(strconv.Itoa(d.Index))
			sb.WriteString("]: ")
			switch d.ChangeType {
			case ChangeTypeAdded:
				fmt.Fprint(&sb, d.Right)
			case ChangeTypeRemoved:
				fmt.Fprint(&sb, d.Left)
			default:
				fmt.Fprint(&sb, d.Left)
				sb.WriteString(" -> ")
				fmt.Fprint(&sb, d.Right)
			}
			sb.WriteString("\n")
		case *StructDiff:
			sb.WriteString(string(d.ChangeType))
			if d.FieldName != "" {
				sb.WriteString(" ")
				if d.Path == d.FieldName {
					sb.WriteString(d.FieldName)
					sb.WriteString(": ")
				} else {
					pathParts := strings.Split(d.Path, ".")
					if len(pathParts) > 1 && pathParts[len(pathParts)-1] == d.FieldName {
						parentPath := strings.Join(pathParts[:len(pathParts)-1], ".")
						if parentPath == "" {
							sb.WriteString(d.FieldName)
							sb.WriteString(": ")
						} else {
							sb.WriteString(parentPath)
							sb.WriteString(".")
							sb.WriteString(d.FieldName)
							sb.WriteString(": ")
						}
					} else {
						sb.WriteString(d.Path)
						sb.WriteString(": ")
					}
				}
			} else {
				sb.WriteString(": ")
			}
			switch d.ChangeType {
			case ChangeTypeAdded:
				fmt.Fprint(&sb, d.Right)
			case ChangeTypeRemoved:
				fmt.Fprint(&sb, d.Left)
			default:
				fmt.Fprint(&sb, d.Left)
				sb.WriteString(" -> ")
				fmt.Fprint(&sb, d.Right)
			}
			sb.WriteString("\n")
		case *Diff:
			sb.WriteString("UPDATED ")
			sb.WriteString(d.Path)
			sb.WriteString(": ")
			fmt.Fprint(&sb, d.Left)
			sb.WriteString(" -> ")
			fmt.Fprint(&sb, d.Right)
			sb.WriteString("\n")
		default:
			sb.WriteString("? Unknown diff type\n")
		}
	}

	return sb.String()
}

// HasDifferences returns true if there are any differences
func (dr *DiffResult) HasDifferences() bool {
	return dr != nil && len(dr.Diffs) > 0
}

// Count returns the number of differences
func (dr *DiffResult) Count() int {
	if dr == nil {
		return 0
	}
	return len(dr.Diffs)
}

type jsonChange struct {
	Type      string  `json:"type"`
	Path      string  `json:"path"`
	Left      any     `json:"leftValue,omitzero"`
	Right     any     `json:"rightValue,omitzero"`
	Key       *string `json:"key,omitzero"`
	Index     *int    `json:"index,omitzero"`
	FieldName string  `json:"fieldName,omitempty"`
	Change    string  `json:"change"`
}

// MarshalJSON implements json.Marshaler using the same change-oriented schema
// as ToJSON. In particular, slice index zero and empty map keys are retained.
func (dr *DiffResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(dr.jsonChanges(), json.Deterministic(true))
}

// MarshalJSONTo implements json.MarshalerTo without allocating an intermediate
// JSON buffer. MarshalEncode preserves any options configured by the caller.
func (dr *DiffResult) MarshalJSONTo(out *jsontext.Encoder) error {
	return json.MarshalEncode(out, dr.jsonChanges())
}

func (dr *DiffResult) jsonChanges() []jsonChange {
	count := 0
	if dr != nil {
		count = len(dr.Diffs)
	}
	changes := make([]jsonChange, 0, count)
	if dr == nil {
		return changes
	}

	for _, diff := range dr.Diffs {
		var jc jsonChange
		switch d := diff.(type) {
		case *MapDiff:
			key := fmt.Sprintf("%v", d.Key)
			jc = jsonChange{
				Type:   "map",
				Path:   d.Path,
				Left:   d.Left,
				Right:  d.Right,
				Key:    &key,
				Change: string(d.ChangeType),
			}
		case *SliceDiff:
			index := d.Index
			jc = jsonChange{
				Type:   "slice",
				Path:   d.Path,
				Left:   d.Left,
				Right:  d.Right,
				Index:  &index,
				Change: string(d.ChangeType),
			}
		case *StructDiff:
			parentPath := d.Path
			if d.FieldName != "" {
				pathParts := strings.Split(d.Path, ".")
				if len(pathParts) > 1 && pathParts[len(pathParts)-1] == d.FieldName {
					parentPath = strings.Join(pathParts[:len(pathParts)-1], ".")
				} else if d.Path == d.FieldName {
					parentPath = ""
				}
			}
			jc = jsonChange{
				Type:      "struct",
				Path:      parentPath,
				Left:      d.Left,
				Right:     d.Right,
				FieldName: d.FieldName,
				Change:    string(d.ChangeType),
			}
		case *Diff:
			jc = jsonChange{Type: "value", Path: d.Path, Left: d.Left, Right: d.Right, Change: "UPDATED"}
		default:
			jc = jsonChange{Type: "unknown", Path: "unknown", Change: "UNKNOWN"}
		}
		changes = append(changes, jc)
	}
	return changes
}

// ToJSON returns a JSON representation of the diff result
func (dr *DiffResult) ToJSON() string {
	if dr == nil || len(dr.Diffs) == 0 {
		return `[]`
	}

	jsonBytes, err := json.Marshal(
		dr,
		json.Deterministic(true),
		jsontext.Multiline(true),
		jsontext.WithIndent("  "),
	)
	if err != nil {
		errorJSON, _ := json.Marshal(
			[]map[string]string{{"error": "Failed to marshal JSON: " + err.Error()}},
			json.Deterministic(true),
		)
		return string(errorJSON)
	}

	return string(jsonBytes)
}

// String returns a human-readable representation of the ChangeType
func (ct ChangeType) String() string {
	switch ct {
	case ChangeTypeAdded:
		return "added"
	case ChangeTypeRemoved:
		return "removed"
	case ChangeTypeUpdated:
		return "updated"
	default:
		return string(ct)
	}
}
