package linescript6

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"
)

// ToJsonF serializes data to JSON with pretty-printing and recursion handling.
// If it detects a recursive object or slice/list, will insert a reference string instead of recursing infinitely.
func ToJsonF(v any) string {
	visited := make(map[uintptr]RecInfo)
	parentChain := []uintptr{}
	return toJsonFInternalNav(v, "", visited, "", parentChain, 0)
}

// toJsonFInternalNav is the enhanced impl, tracking parent pointers and emitting "this.parent[...].foo.bar..." for cycles.
type RecInfo struct {
	Path   string
	Parent uintptr
	Level  int
}

func toJsonFInternalNav(v any, path string, visited map[uintptr]RecInfo, indent string, parentChain []uintptr, level int) string {
	const indentStep = "    "
	switch vv := v.(type) {
	case *Record:
		ptr := uintptrPointer(vv)
		if ptr != 0 {
			if prev, ok := visited[ptr]; ok {
				// Found cycle. Reconstruct parent navigation. This level is "level".
				nav := buildNavPath(prev, parentChain)
				return fmt.Sprintf(`"%s"`, nav)
			}
			visInfo := struct {
				Path   string
				Parent uintptr
				Level  int
			}{
				Path:   path,
				Parent: 0,
				Level:  level,
			}
			if len(parentChain) > 0 {
				visInfo.Parent = parentChain[len(parentChain)-1]
			}
			visited[ptr] = visInfo
		}
		var sb strings.Builder
		sb.WriteString("{\n")
		nextIndent := indent + indentStep
		for i, k := range vv.Keys {
			if i != 0 {
				sb.WriteString(",\n")
			}
			sb.WriteString(nextIndent)
			sb.WriteString(fmt.Sprintf(`"%s": `, k))
			childPath := path + "." + k
			childParents := append(parentChain, ptr)
			sb.WriteString(toJsonFInternalNav(vv.Values[i], childPath, visited, nextIndent, childParents, level+1))
		}
		sb.WriteString("\n")
		sb.WriteString(indent)
		sb.WriteString("}")
		return sb.String()
	case *List:
		ptr := uintptrPointer(vv)
		if ptr != 0 {
			if prev, ok := visited[ptr]; ok {
				nav := buildNavPath(prev, parentChain)
				return fmt.Sprintf(`"%s"`, nav)
			}
			visInfo := struct {
				Path   string
				Parent uintptr
				Level  int
			}{
				Path:   path,
				Parent: 0,
				Level:  level,
			}
			if len(parentChain) > 0 {
				visInfo.Parent = parentChain[len(parentChain)-1]
			}
			visited[ptr] = visInfo
		}
		var sb strings.Builder
		sb.WriteString("[\n")
		nextIndent := indent + indentStep
		for i, el := range vv.TheSlice {
			if i != 0 {
				sb.WriteString(",\n")
			}
			sb.WriteString(nextIndent)
			childPath := fmt.Sprintf("%s[%d]", path, i)
			childParents := append(parentChain, ptr)
			sb.WriteString(toJsonFInternalNav(el, childPath, visited, nextIndent, childParents, level+1))
		}
		sb.WriteString("\n")
		sb.WriteString(indent)
		sb.WriteString("]")
		return sb.String()
	case string:
		b, _ := json.Marshal(vv)
		return string(b)
	case nil:
		return "null"
	case int, int64, float64, bool:
		b, _ := json.Marshal(vv)
		return string(b)
	default:
		b, _ := json.MarshalIndent(vv, "", "    ")
		return string(b)
	}
}

// buildNavPath builds a navigation path like "this.parent.parent.foo[0].bar"
func buildNavPath(info struct {
	Path   string
	Parent uintptr
	Level  int
}, parentChain []uintptr) string {
	// Walk up parentChain to info.Level
	nParent := len(parentChain) - info.Level
	sb := strings.Builder{}
	sb.WriteString("this")
	for i := 0; i < nParent; i++ {
		sb.WriteString(".parent")
	}
	// info.Path starts with ".", so add if present but chop leading "."
	if len(info.Path) > 0 {
		if info.Path[0] == '.' {
			sb.WriteString(info.Path)
		} else {
			sb.WriteString(".")
			sb.WriteString(info.Path)
		}
	}
	return sb.String()
}

// uintptrPointer returns the uintptr of a pointer type, 0 otherwise
func uintptrPointer(v any) uintptr {
	switch vv := v.(type) {
	case *Record:
		return uintptrOf(vv)
	case *List:
		return uintptrOf(vv)
	default:
		return 0
	}
}
func uintptrOf(p any) uintptr {
	return uintptr((*[2]uintptr)(unsafe.Pointer(&p))[1])
}
