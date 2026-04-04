package scheduler

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockJob struct {
	runFunc func() error
}

func (m *mockJob) Run() error {
	if m.runFunc != nil {
		return m.runFunc()
	}
	return nil
}

func TestNewScheduler(t *testing.T) {
	s := NewScheduler(4, 100)

	assert.Equal(t, 4, s.workers)
	assert.NotNil(t, s.jobs)
	assert.NotNil(t, s.stopChan)
}

func TestScheduler_StartAndStop(t *testing.T) {
	s := NewScheduler(2, 10)
	s.Start()
	s.Stop()
}

func TestScheduler_EnqueueAndProcess(t *testing.T) {
	s := NewScheduler(1, 10)
	s.Start()

	var wg sync.WaitGroup
	wg.Add(1)

	job := &mockJob{
		runFunc: func() error {
			wg.Done()
			return nil
		},
	}

	s.Enqueue(job)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job was not processed in time")
	}

	s.Stop()
}

func TestScheduler_JobError(t *testing.T) {
	s := NewScheduler(1, 10)
	s.Start()

	var wg sync.WaitGroup
	wg.Add(1)

	job := &mockJob{
		runFunc: func() error {
			wg.Done()
			return errors.New("job error")
		},
	}

	s.Enqueue(job)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job was not processed in time")
	}

	s.Stop()
}

func TestScheduler_QueueFull(t *testing.T) {
	s := NewScheduler(1, 2)
	s.Start()

	var blockJob sync.WaitGroup
	blockJob.Add(1)

	blockingJob := &mockJob{
		runFunc: func() error {
			blockJob.Wait()
			return nil
		},
	}

	s.Enqueue(blockingJob)
	time.Sleep(50 * time.Millisecond)

	s.Enqueue(&mockJob{})
	s.Enqueue(&mockJob{})
	s.Enqueue(&mockJob{})

	blockJob.Done()
	time.Sleep(100 * time.Millisecond)

	s.Stop()
}

func TestScheduler_MultipleWorkers(t *testing.T) {
	s := NewScheduler(4, 100)
	s.Start()

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		job := &mockJob{
			runFunc: func() error {
				wg.Done()
				return nil
			},
		}
		s.Enqueue(job)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("not all jobs were processed in time")
	}

	s.Stop()
}
