package tomledit

// Field is one ordered field in an array-of-tables entry.
type Field struct {
	Key   string
	Value Value
}

// F creates an ordered table field.
func F(key string, value Value) Field {
	return Field{Key: key, Value: value}
}
