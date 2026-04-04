package scheduler

import (
	"log"
	"sync"
)

type Job interface {
	Run() error
}

type Scheduler struct {
	jobs     chan Job
	workers  int
	wg       sync.WaitGroup
	stopChan chan struct{}
}

func NewScheduler(workers, queueSize int) *Scheduler {
	return &Scheduler{
		jobs:     make(chan Job, queueSize),
		workers:  workers,
		stopChan: make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	log.Printf("Scheduler started with %d workers", s.workers)
}

func (s *Scheduler) worker(id int) {
	defer s.wg.Done()
	for {
		select {
		case job := <-s.jobs:
			if err := job.Run(); err != nil {
				log.Printf("Worker %d: job failed: %v", id, err)
			}
		case <-s.stopChan:
			log.Printf("Worker %d: stopping", id)
			return
		}
	}
}

func (s *Scheduler) Enqueue(job Job) {
	select {
	case s.jobs <- job:
	default:
		log.Println("Job queue is full, dropping job")
	}
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Scheduler stopped")
}
