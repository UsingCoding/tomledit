package tomledit

type valueKind uint8

const (
	valueString valueKind = iota + 1
	valueInteger
	valueFloat
	valueBoolean
	valueArray
	valueRaw
)

// Value is a TOML value supplied to an Operation.
type Value struct {
	kind    valueKind
	string  string
	integer int64
	float   float64
	boolean bool
	array   []Value
}

// String creates a TOML basic string value.
func String(v string) Value {
	return Value{kind: valueString, string: v}
}

// Integer creates a TOML integer value.
func Integer(v int64) Value {
	return Value{kind: valueInteger, integer: v}
}

// Float creates a TOML floating-point value. Non-finite values are rejected by Apply.
func Float(v float64) Value {
	return Value{kind: valueFloat, float: v}
}

// Boolean creates a TOML boolean value.
func Boolean(v bool) Value {
	return Value{kind: valueBoolean, boolean: v}
}

// Array creates a TOML array value.
func Array(values ...Value) Value {
	return Value{kind: valueArray, array: append([]Value(nil), values...)}
}

// Raw parses v as a TOML value before inserting it.
func Raw(v string) Value {
	return Value{kind: valueRaw, string: v}
}
