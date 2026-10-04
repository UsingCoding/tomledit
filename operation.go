package tomledit

type operationKind uint8

const (
	opSet operationKind = iota + 1
	opDelete
	opArrayAppend
	opArrayInsert
	opArrayReplace
	opArrayRemove
	opAOTAppend
	opAOTInsert
	opAOTReplace
	opAOTRemove
)

// Operation is a TOML mutation to be applied by Editor.Apply.
type Operation struct {
	kind   operationKind
	path   Path
	value  Value
	index  int
	fields []Field
}

// Set inserts or replaces the terminal key at path.
func Set(path Path, value Value) Operation {
	return Operation{kind: opSet, path: clonePath(path), value: value}
}

// Delete removes the terminal key at path.
func Delete(path Path) Operation {
	return Operation{kind: opDelete, path: clonePath(path)}
}

// ArrayAppend appends value to the array at path.
func ArrayAppend(path Path, value Value) Operation {
	return Operation{kind: opArrayAppend, path: clonePath(path), value: value}
}

// ArrayInsert inserts value at index in the array at path.
func ArrayInsert(path Path, index int, value Value) Operation {
	return Operation{kind: opArrayInsert, path: clonePath(path), index: index, value: value}
}

// ArrayReplace replaces the value at index in the array at path.
func ArrayReplace(path Path, index int, value Value) Operation {
	return Operation{kind: opArrayReplace, path: clonePath(path), index: index, value: value}
}

// ArrayRemove removes the value at index in the array at path.
func ArrayRemove(path Path, index int) Operation {
	return Operation{kind: opArrayRemove, path: clonePath(path), index: index}
}

// AOTAppend appends an entry to the array of tables at path.
func AOTAppend(path Path, fields ...Field) Operation {
	return Operation{kind: opAOTAppend, path: clonePath(path), fields: append([]Field(nil), fields...)}
}

// AOTInsert inserts an entry at index in the array of tables at path.
func AOTInsert(path Path, index int, fields ...Field) Operation {
	return Operation{kind: opAOTInsert, path: clonePath(path), index: index, fields: append([]Field(nil), fields...)}
}

// AOTReplace replaces the entry at index in the array of tables at path.
func AOTReplace(path Path, index int, fields ...Field) Operation {
	return Operation{kind: opAOTReplace, path: clonePath(path), index: index, fields: append([]Field(nil), fields...)}
}

// AOTRemove removes the entry at index in the array of tables at path.
func AOTRemove(path Path, index int) Operation {
	return Operation{kind: opAOTRemove, path: clonePath(path), index: index}
}
