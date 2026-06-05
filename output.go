package godiff

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// String returns a human-readable representation of the diff result
func (dr *DiffResult) String() string {
	if len(dr.Diffs) == 0 {
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
	return len(dr.Diffs) > 0
}

// Count returns the number of differences
func (dr *DiffResult) Count() int {
	return len(dr.Diffs)
}

// ToJSON returns a JSON representation of the diff result
func (dr *DiffResult) ToJSON() string {
	if len(dr.Diffs) == 0 {
		return `[]`
	}

	type jsonChange struct {
		Type      string `json:"type"`
		Path      string `json:"path"`
		Left      any    `json:"leftValue,omitempty"`
		Right     any    `json:"rightValue,omitempty"`
		Key       string `json:"key,omitempty"`
		Index     int    `json:"index,omitempty"`
		FieldName string `json:"fieldName,omitempty"`
		Change    string `json:"change"`
	}

	changes := make([]jsonChange, 0, len(dr.Diffs))

	for _, diff := range dr.Diffs {
		var jc jsonChange
		switch d := diff.(type) {
		case *MapDiff:
			jc = jsonChange{
				Type:   "map",
				Path:   d.Path,
				Left:   d.Left,
				Right:  d.Right,
				Key:    fmt.Sprintf("%v", d.Key),
				Change: string(d.ChangeType),
			}
		case *SliceDiff:
			jc = jsonChange{
				Type:   "slice",
				Path:   d.Path,
				Left:   d.Left,
				Right:  d.Right,
				Index:  d.Index,
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
			jc = jsonChange{
				Type:   "unknown",
				Path:   "unknown",
				Left:   nil,
				Right:  nil,
				Change: "UNKNOWN",
			}
		}
		changes = append(changes, jc)
	}

	jsonBytes, err := json.MarshalIndent(changes, "", "  ")
	if err != nil {
		return fmt.Sprintf(`[{"error": "Failed to marshal JSON: %s"}]`, err.Error())
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
