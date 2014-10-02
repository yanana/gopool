package gopool

import "fmt"

type Pool struct {
	PoolConfig
	TaskQueue        chan Task
	WorkerQueue      chan chan Task
	NumOfQueuedTasks int
	Semaphore        chan int
}

type PoolConfig struct {
	PoolSize      int
	QueueCapacity int
	Debug         bool
}

func (p *Pool) StartDispatcher() {
	p.log("Pool.PoolSize: %d\n", p.PoolSize)
	for i := 0; i < p.PoolSize; i++ {
		p.log("Starting worker[%d]\n", i)
		worker := NewWorker(i+1, p.WorkerQueue)
		worker.Start(p)
	}

	go func() {
		for {
			select {
			case task := <-p.TaskQueue:
				p.log("Received task")
				go func() {
					worker := <-p.WorkerQueue
					p.log("Dispatching task")
					worker <- task
				}()
			}
		}
	}()
}

func NewPool(poolSize int, queueCapacity int, debug bool) *Pool {
	semaphore := make(chan int, queueCapacity)
	for i := 0; i < queueCapacity; i++ {
		semaphore <- 1
	}
	pool := Pool{
		TaskQueue:   make(chan Task, queueCapacity),
		WorkerQueue: make(chan chan Task, poolSize),
		PoolConfig: PoolConfig{
			PoolSize:      poolSize,
			QueueCapacity: queueCapacity,
			Debug:         debug,
		},
		NumOfQueuedTasks: 0,
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
				pool.log("worker[%d]: Received task\n", w.Id)
				task.Execute()
				pool.Semaphore <- 1
			case <-w.ShutdownChan:
				pool.log("worker[%d] stopping\n", w.Id)
			}
		}
	}()
}

type Task interface {
	Execute()
}

func (p *Pool) log(format string, a ...interface{}) {
	if p.Debug {
		fmt.Printf(format, a)
	}
}
