package gopool

import "fmt"

type Pool struct {
	taskQueue        chan Task
	workerQueue      chan chan Task
	poolSize         int
	queueCapacity    int
	numOfQueuedTasks int
	semaphore        chan int
}

func (p *Pool) StartDispatcher() {
	fmt.Printf("Pool.poolSize: %d\n", p.poolSize)
	for i := 0; i < p.poolSize; i++ {
		fmt.Println("Starting worker[%d]", i)
		worker := NewWorker(i+1, p.workerQueue)
		worker.Start(p)
	}

	go func() {
		for {
			select {
			case task := <-p.taskQueue:
				fmt.Println("Received task")
				go func() {
					worker := <-p.workerQueue
					fmt.Println("Dispatching task")
					worker <- task
				}()
			}
		}
	}()
}

func NewPool(poolSize int, queueCapacity int) *Pool {
	semaphore := make(chan int, queueCapacity)
	for i := 0; i < queueCapacity; i++ {
		semaphore <- 1
	}
	pool := Pool{
		taskQueue:        make(chan Task, queueCapacity),
		workerQueue:      make(chan chan Task, poolSize),
		poolSize:         poolSize,
		numOfQueuedTasks: 0,
		queueCapacity:    queueCapacity,
		semaphore:        semaphore,
	}
	return &pool
}

func (pool *Pool) Submit(task Task) {
	pool.taskQueue <- task
}

type Worker struct {
	Id           int
	Task         chan Task
	WorkerQueue  chan chan Task
	ShutdownChan chan bool
}

func NewWorker(id int, workerQueue chan chan Task) Worker {
	worker := Worker{
		Id:           id,
		Task:         make(chan Task),
		WorkerQueue:  workerQueue,
		ShutdownChan: make(chan bool),
	}

	return worker
}

func (w Worker) Start(pool *Pool) {
	go func() {
		for {
			w.WorkerQueue <- w.Task
			select {
			case task := <-w.Task:
				fmt.Printf("worker[%d]: Received task\n", w.Id)
				task.Execute()
				pool.semaphore <- 1
			case <-w.ShutdownChan:
				fmt.Printf("worker[%d] stopping\n", w.Id)
			}
		}
	}()
}

type Task interface {
	Execute()
}
