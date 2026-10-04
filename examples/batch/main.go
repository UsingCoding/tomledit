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

	output, err := editor.Apply(context.Background(), []byte("version = \"1\"\nplugins = [\"git\"]\n"),
		tomledit.Set(tomledit.P(tomledit.K("version")), tomledit.String("2")),
		tomledit.ArrayAppend(tomledit.P(tomledit.K("plugins")), tomledit.String("golang")),
	)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
