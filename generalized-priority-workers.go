package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func priorityWorker(priority, totalWorkerKinds int, priorityMutexes []sync.Mutex,
	cntMutex *sync.Mutex, cntWorker []int,
	wg, beginBarrier *sync.WaitGroup) {

	beginBarrier.Done()
	beginBarrier.Wait()

	for i := priority; i < totalWorkerKinds; i++ {
		priorityMutexes[i].Lock()
	}
	for i := totalWorkerKinds - 1; i >= priority; i-- {
		priorityMutexes[i].Unlock()
	}

	cntMutex.Lock()
	cntWorker[priority]++
	if cntWorker[priority] == 1 {
		for i := 0; i < priority; i++ {
			priorityMutexes[i].Lock()
		}
	}
	cntMutex.Unlock()

	// Critical section
	fmt.Println("Priority of the worker:", priority)

	cntMutex.Lock()
	cntWorker[priority]--
	if cntWorker[priority] == 0 {
		cntMutex.Unlock()
		for i := priority - 1; i >= 0; i-- {
			priorityMutexes[i].Unlock()
		}
	} else {
		cntMutex.Unlock()
	}

	wg.Done()
}

func main() {
	const totalWorkerKinds = 100
	cntWorker := make([]int, totalWorkerKinds)
	priorityMutexes := make([]sync.Mutex, totalWorkerKinds)
	cntMutex := make([]sync.Mutex, totalWorkerKinds)

	const numWorkers = 1000
	var wg, beginBarrier sync.WaitGroup
	wg.Add(numWorkers)
	beginBarrier.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		priority := rand.Int() % totalWorkerKinds
		go priorityWorker(priority, totalWorkerKinds, priorityMutexes,
			&cntMutex[priority], cntWorker, &wg, &beginBarrier)
	}

	wg.Wait()
}
