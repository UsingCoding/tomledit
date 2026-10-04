package tomledit

// Path identifies a TOML location with unambiguous key and index elements.
type Path []PathElement

type pathElementKind uint8

const (
	pathKey pathElementKind = iota + 1
	pathIndex
)

// PathElement is one key or array index in a Path.
type PathElement struct {
	kind  pathElementKind
	key   string
	index int
}

// P creates a Path from typed elements.
func P(parts ...PathElement) Path {
	return append(Path(nil), parts...)
}

// K creates a key path element.
func K(v string) PathElement {
	return PathElement{kind: pathKey, key: v}
}

// I creates an array-index path element.
func I(v int) PathElement {
	return PathElement{kind: pathIndex, index: v}
}

func clonePath(path Path) Path {
	return append(Path(nil), path...)
}
