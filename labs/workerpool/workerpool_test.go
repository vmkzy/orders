package workerpool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func BenchmarkRun(b *testing.B) {
	jobs := make([]func(context.Context) error, 64)
	for i := range jobs {
		jobs[i] = func(ctx context.Context) error {
			time.Sleep(time.Millisecond)
			return nil
		}
	}
	test := []int{1, 4, 16, 64}
	for _, n := range test {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := Run(context.Background(), n, jobs); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
