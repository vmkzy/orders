package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

func Run(ctx context.Context, n int, jobs []func(context.Context) error) error {
	if n <= 0 {
		return fmt.Errorf("invalid n")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(jobs) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	list := make(chan int)
	sliceError := make([]error, len(jobs))

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for indx := range list {
				if ctx.Err() != nil {
					return
				}
				if err := jobs[indx](ctx); err != nil {
					sliceError[indx] = err
				}

			}
		}()
	}

loop:
	for id := range jobs {
		if ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
			break loop
		case list <- id:
		}
	}
	close(list)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return err
	}
	return errors.Join(sliceError...)
}
