package gopool

import "fmt"

type Pool struct {
	TaskQueue        chan Task
	WorkerQueue      chan chan Task
	PoolSize         int
	QueueCapacity    int
	NumOfQueuedTasks int
	Semaphore        chan int
}

func (p *Pool) StartDispatcher() {
	fmt.Printf("Pool.poolSize: %d\n", p.PoolSize)
	for i := 0; i < p.PoolSize; i++ {
		fmt.Println("Starting worker[%d]", i)
		worker := NewWorker(i+1, p.WorkerQueue)
		worker.Start(p)
	}

	go func() {
		for {
			select {
			case task := <-p.TaskQueue:
				fmt.Println("Received task")
				go func() {
					worker := <-p.WorkerQueue
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
		TaskQueue:        make(chan Task, queueCapacity),
		WorkerQueue:      make(chan chan Task, poolSize),
		PoolSize:         poolSize,
		NumOfQueuedTasks: 0,
		QueueCapacity:    queueCapacity,
		Semaphore:        semaphore,
	}
	return &pool
}

func (pool *Pool) Submit(task Task) {
	pool.TaskQueue <- task
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
				pool.Semaphore <- 1
			case <-w.ShutdownChan:
				fmt.Printf("worker[%d] stopping\n", w.Id)
			}
		}
	}()
}

type Task interface {
	Execute()
}
