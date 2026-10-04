package tomledit_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/usingcoding/tomledit"
)

func BenchmarkApply(b *testing.B) {
	for _, size := range []int{10_000, 100_000, 1_000_000} {
		for _, operationCount := range []int{1, 10} {
			b.Run(fmt.Sprintf("%dB/%d-operations", size, operationCount), func(b *testing.B) {
				editor, err := tomledit.New(context.Background())
				if err != nil {
					b.Fatal(err)
				}
				defer editor.Close(context.Background())
				input := benchmarkDocument(size)
				operations := make([]tomledit.Operation, operationCount)
				for i := range operations {
					operations[i] = tomledit.Set(tomledit.P(tomledit.K("version")), tomledit.String(fmt.Sprintf("%d", i)))
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, err := editor.Apply(context.Background(), input, operations...); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func benchmarkDocument(size int) []byte {
	padding := strings.Repeat("# benchmark padding\n", size/20+1)
	return []byte(padding + "version = \"initial\"\n")
}
