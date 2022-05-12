package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ActionType int

const (
	None  ActionType = -1
	Read  ActionType = 0
	Write ActionType = 1
)

const NumActionType = 2

func ActionName(actionType ActionType) string {
	switch actionType {
	case None:
		return "None"
	case Read:
		return "Read"
	case Write:
		return "Write"
	default:
		panic("unrecognized action")
	}
}

type WorkerInfo struct {
	actionType ActionType
	ackChan    chan int
}

func Worker(requestChan chan WorkerInfo, finishChan chan int, actionType ActionType,
	num int, wg *sync.WaitGroup) {

	request := WorkerInfo{actionType: actionType, ackChan: make(chan int, 1)}
	requestChan <- request

	<-request.ackChan

	// critical region begins
	fmt.Printf("%s %d begins\n", ActionName(actionType), num)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("%s %d ends\n", ActionName(actionType), num)
	// critical region ends

	finishChan <- 0 // this can be any integer
	wg.Done()
}

func Arbitrator(requestChan chan WorkerInfo, finishChan chan int, typeCap map[ActionType]int) {
	var workerCnt [NumActionType]int
	curActionType := None
	diffTypeNext := false         // Whether the next worker awaiting belongs to another type
	var nextWorkerInfo WorkerInfo // valid when diffTypeNext == true

	for {
		if curActionType == None {
			request := <-requestChan
			curActionType = request.actionType
			workerCnt[curActionType]++
			request.ackChan <- 0 // this can be any integer
		} else if workerCnt[curActionType] < typeCap[curActionType] && diffTypeNext == false {
			select {
			case request := <-requestChan:
				if request.actionType == curActionType {
					workerCnt[curActionType]++
					request.ackChan <- 0 // this can be any integer
				} else {
					diffTypeNext = true
					nextWorkerInfo = request
				}
			case <-finishChan:
				workerCnt[curActionType]--
				if workerCnt[curActionType] == 0 {
					if diffTypeNext {
						curActionType = nextWorkerInfo.actionType
						diffTypeNext = false
						workerCnt[curActionType]++
						nextWorkerInfo.ackChan <- 0 // this can be any integer
					} else {
						curActionType = None
					}
				}
			}
		} else {
			<-finishChan
			workerCnt[curActionType]--
			if workerCnt[curActionType] == 0 {
				if diffTypeNext {
					curActionType = nextWorkerInfo.actionType
					diffTypeNext = false
					workerCnt[curActionType]++
					nextWorkerInfo.ackChan <- 0 // this can be any integer
				} else {
					curActionType = None
				}
			}
		}
	}
}

func ArbitratorSecond(requestChan chan WorkerInfo, finishChan chan int, typeCap map[ActionType]int) {
	var workerCnt [NumActionType]int
	curActionType := None

	for {
		if curActionType != None && workerCnt[curActionType] == typeCap[curActionType] {
			<-finishChan
			workerCnt[curActionType]--
		} else {
			request := <-requestChan
			if request.actionType != curActionType {
				if curActionType != None {
					for i := 0; i < workerCnt[curActionType]; i++ {
						<-finishChan
					}
					workerCnt[curActionType] = 0
				}
				curActionType = request.actionType
			}
			workerCnt[curActionType]++
			request.ackChan <- 0 // this can be any integer
		}
	}
}

func main() {
	const infiniteCap = 10000
	requestChan := make(chan WorkerInfo, infiniteCap)
	finishChan := make(chan int, infiniteCap)
	var typeCap = map[ActionType]int{
		Read:  infiniteCap,
		Write: 1,
	}

	go ArbitratorSecond(requestChan, finishChan, typeCap)

	const numWorkers = 100
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go Worker(requestChan, finishChan, ActionType(rand.Int()%NumActionType), i, &wg)
	}

	wg.Wait()
}
