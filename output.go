package godiff

import (
	"encoding/json"
	"fmt"
	"strings"
)

// String returns a human-readable representation of the diff result
func (dr *DiffResult) String() string {
	if len(dr.Diffs) == 0 {
		return "No differences found"
	}

	var sb strings.Builder
	estimatedSize := len(dr.Diffs) * 80
	for _, diff := range dr.Diffs {
		if d, ok := diff.(*Diff); ok {
			if len(d.Path) > 20 {
				estimatedSize += len(d.Path) * 2
			}
		}
	}
	sb.Grow(estimatedSize)

	sb.WriteString(fmt.Sprintf("Found %d differences:\n", len(dr.Diffs)))

	for _, diff := range dr.Diffs {
		// Use type assertion to handle different diff types
		switch d := diff.(type) {
		case *MapDiff:
			sb.WriteString(string(d.ChangeType))
			sb.WriteString(" ")
			sb.WriteString(d.Path)
			sb.WriteString(fmt.Sprintf("[%s]: ", d.Key))
			switch d.ChangeType {
			case ChangeTypeAdded:
				sb.WriteString(fmt.Sprintf("%v", d.Right))
			case ChangeTypeRemoved:
				sb.WriteString(fmt.Sprintf("%v", d.Left))
			default:
				sb.WriteString(fmt.Sprintf("%v -> %v", d.Left, d.Right))
			}
			sb.WriteString("\n")
		case *SliceDiff:
			sb.WriteString(string(d.ChangeType))
			sb.WriteString(" ")
			sb.WriteString(d.Path)
			sb.WriteString(fmt.Sprintf("[%d]: ", d.Index))
			switch d.ChangeType {
			case ChangeTypeAdded:
				sb.WriteString(fmt.Sprintf("%v", d.Right))
			case ChangeTypeRemoved:
				sb.WriteString(fmt.Sprintf("%v", d.Left))
			default:
				sb.WriteString(fmt.Sprintf("%v -> %v", d.Left, d.Right))
			}
			sb.WriteString("\n")
		case *StructDiff:
			sb.WriteString(string(d.ChangeType))
			if d.FieldName != "" {
				sb.WriteString(" ")
				if d.Path == d.FieldName {
					sb.WriteString(fmt.Sprintf("%s: ", d.FieldName))
				} else {
					pathParts := strings.Split(d.Path, ".")
					if len(pathParts) > 1 && pathParts[len(pathParts)-1] == d.FieldName {
						parentPath := strings.Join(pathParts[:len(pathParts)-1], ".")
						if parentPath == "" {
							sb.WriteString(fmt.Sprintf("%s: ", d.FieldName))
						} else {
							sb.WriteString(fmt.Sprintf("%s.%s: ", parentPath, d.FieldName))
						}
					} else {
						sb.WriteString(fmt.Sprintf("%s: ", d.Path))
					}
				}
			} else {
				sb.WriteString(": ")
			}
			switch d.ChangeType {
			case ChangeTypeAdded:
				sb.WriteString(fmt.Sprintf("%v", d.Right))
			case ChangeTypeRemoved:
				sb.WriteString(fmt.Sprintf("%v", d.Left))
			default:
				sb.WriteString(fmt.Sprintf("%v -> %v", d.Left, d.Right))
			}
			sb.WriteString("\n")
		case *Diff:
			sb.WriteString("UPDATED ")
			sb.WriteString(d.Path)
			sb.WriteString(": ")
			sb.WriteString(fmt.Sprintf("%v", d.Left))
			sb.WriteString(" -> ")
			sb.WriteString(fmt.Sprintf("%v", d.Right))
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
		// no-op for summary counts anymore
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
	case ChangeTypeIDMismatch:
		return "id mismatch"
	default:
		return string(ct)
	}
}
