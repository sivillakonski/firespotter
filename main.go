package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sivillakonski/firespotter/dto"
	"github.com/sivillakonski/firespotter/plowjob"
)

var (
	fireSpotterCommanderAddress     = flag.String("commander-address", "http://127.0.0.1:8080", "The HTTP address of the command node)")
	fireSpotterCommanderRefreshRate = flag.Duration("command-obey-refresh-rate", time.Minute, "How often to reload commanders orders")
	fireSpotterCommanderFirstTopN   = flag.Int("command-first-top-n", 10, "First top N targets  to process ('-1' stands for all jobs on the list)")
	fireSpotterConnectionsPerTarget = flag.Int("command-connections", 1, "Number of parallel connections per target")
)

func init() {
	flag.Parse()
}

func main() {
	log.Println("Fire spotter is igniting, parameters:")
	log.Printf("\tcommand center: %s\n", *fireSpotterCommanderAddress)
	log.Printf("\tcommand refresh rate: %s\n", fireSpotterCommanderRefreshRate.String())
	log.Printf("\tfirst top N targets to process: %d\n", *fireSpotterCommanderFirstTopN)
	log.Printf("\tconnections per taget host: %d\n", *fireSpotterConnectionsPerTarget)

	orders := make(chan []dto.Target, 128)
	go receiveOrders(orders)

	var (
		lastOrderHash   = ""
		lastOrderCancel func()
	)

	for currentOrder := range orders {
		newOrderHash := fmt.Sprintf("%+v", currentOrder) // TODO: calculate hash based on values

		if newOrderHash == lastOrderHash {
			continue
		}

		log.Printf("-> new order received: %s\n", newOrderHash)
		lastOrderHash = newOrderHash

		if lastOrderCancel != nil {
			lastOrderCancel()
		}

		workersNumber := *fireSpotterCommanderFirstTopN

		// If unlimited - set to the max CPU Cores number
		if workersNumber < 1 {
			workersNumber = runtime.NumCPU()
		}

		// Number of orders could not be more than number of orders
		if len(currentOrder) < workersNumber {
			workersNumber = len(currentOrder)
		}

		queue := NewJobQueue(workersNumber)
		queue.Start()
		lastOrderCancel = queue.Stop

		for i := 0; i < workersNumber; i++ {
			queue.Submit(plowjob.NewPlowJob(currentOrder[i].URL, *fireSpotterConnectionsPerTarget))
		}
	}
}

func receiveOrders(orders chan []dto.Target) {
	client := resty.New()
	ticker := time.NewTicker(*fireSpotterCommanderRefreshRate)

	for {
		targets := make([]dto.Target, 0)

		resp, err := client.R().
			SetResult(&targets).
			Get(*fireSpotterCommanderAddress)

		if err == nil && resp.IsSuccess() {
			orders <- targets
			<-ticker.C
		} else {
			if err != nil {
				log.Printf("[ERROR] failed to retrieve the new target list: %s\n", err)
			} else {
				log.Printf("[ERROR] failed to retrieve the new target list, response: %s", resp.String())
			}

			time.Sleep(time.Second)
		}
	}
}

// Job - interface for job processing
type Job interface {
	Process()
	Cancel()
	FinishedSignal() chan struct{}
}

// Worker - the worker threads that actually process the jobs
type Worker struct {
	done             sync.WaitGroup
	readyPool        chan chan Job
	assignedJobQueue chan Job

	quit chan bool
}

// JobQueue - a queue for enqueueing jobs to be processed
type JobQueue struct {
	internalQueue     chan Job
	readyPool         chan chan Job
	workers           []*Worker
	dispatcherStopped sync.WaitGroup
	workersStopped    sync.WaitGroup
	quit              chan bool
}

// NewJobQueue - creates a new job queue
func NewJobQueue(maxWorkers int) *JobQueue {
	workersStopped := sync.WaitGroup{}
	readyPool := make(chan chan Job, maxWorkers)
	workers := make([]*Worker, maxWorkers, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = NewWorker(readyPool, workersStopped)
	}
	return &JobQueue{
		internalQueue:     make(chan Job),
		readyPool:         readyPool,
		workers:           workers,
		dispatcherStopped: sync.WaitGroup{},
		workersStopped:    workersStopped,
		quit:              make(chan bool),
	}
}

// Start - starts the worker routines and dispatcher routine
func (q *JobQueue) Start() {
	for i := 0; i < len(q.workers); i++ {
		q.workers[i].Start()
	}
	go q.dispatch()
}

// Stop - stops the workers and sispatcher routine
func (q *JobQueue) Stop() {
	q.quit <- true
	q.dispatcherStopped.Wait()
}

func (q *JobQueue) dispatch() {
	q.dispatcherStopped.Add(1)
	for {
		select {
		case job := <-q.internalQueue: // We got something in on our queue
			workerChannel := <-q.readyPool // Check out an available worker
			workerChannel <- job           // Send the request to the channel
		case <-q.quit:
			for i := 0; i < len(q.workers); i++ {
				q.workers[i].Stop()
			}
			q.workersStopped.Wait()
			q.dispatcherStopped.Done()
			return
		}
	}
}

// Submit - adds a new job to be processed
func (q *JobQueue) Submit(job Job) {
	q.internalQueue <- job
}

// NewWorker - creates a new worker
func NewWorker(readyPool chan chan Job, done sync.WaitGroup) *Worker {
	return &Worker{
		done:             done,
		readyPool:        readyPool,
		assignedJobQueue: make(chan Job),
		quit:             make(chan bool),
	}
}

// Start - begins the job processing loop for the worker
func (w *Worker) Start() {
	go func() {
		w.done.Add(1)
		for {
			w.readyPool <- w.assignedJobQueue // check the job queue in
			select {
			case job := <-w.assignedJobQueue: // see if anything has been assigned to the queue
				go job.Process()

				select {
				case <-job.FinishedSignal():
					// That's fine, ready for a new Job
				case <-w.quit:
					job.Cancel()
				}

			case <-w.quit:
				w.done.Done()
				return
			}
		}
	}()
}

// Stop - stops the worker
func (w *Worker) Stop() {
	w.quit <- true
}
