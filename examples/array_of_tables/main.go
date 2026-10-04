package main

import (
	"context"
	"fmt"

	"github.com/usingcoding/tomledit"
)

func main() {
	editor, err := tomledit.New(context.Background())
	if err != nil {
		panic(err)
	}
	defer editor.Close(context.Background())

	input := []byte("[[foo.items]]\nname = \"foo\"\n\n[[foo.items]]\nname = \"bar\"\n\n[[foo.items]]\nname = \"baz\"\n")
	output, err := editor.Apply(context.Background(), input, tomledit.AOTInsert(
		tomledit.P(tomledit.K("foo"), tomledit.K("items")),
		2,
		tomledit.F("name", tomledit.String("qux")),
		tomledit.F("value", tomledit.String("qux")),
		tomledit.F("group", tomledit.String("foo")),
	))
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
