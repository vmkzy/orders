package stats

import (
	"sync"
	"testing"
)

func TestRecordParallel(t *testing.T) {
	workers := 5
	records := 1000
	s := New()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < records; j++ {
				s.Record("new")
			}
		}()
	}
	close(start)
	wg.Wait()

	if got, want := s.Total(), int64(workers*records); got != want {
		t.Errorf("total() = %d, want = %d", got, want)
	}
}
