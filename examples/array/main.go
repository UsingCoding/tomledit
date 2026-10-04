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

	output, err := editor.Apply(context.Background(), []byte("tools = [\"go\", \"cargo\"]\n"), tomledit.ArrayInsert(tomledit.P(tomledit.K("tools")), 1, tomledit.String("lint")))
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
