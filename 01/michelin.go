package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const N = 5 // try 100_000 for fun
const GOAL = 3

type Action int

const (
	REQUEST Action = iota
	GRANT
	RELEASE
)

type Event struct {
	src  int
	act  Action
	resp chan Event // reply chan for GRANT
}

func fork(id int, ch chan Event) {
	var holder int = -1 // -1 for free
	var queue []Event

	for {
		var e Event
		if len(queue) > 0 && holder == -1 {
			e, queue = queue[0], queue[1:]
		} else {
			e = <-ch
		}

		switch e.act {
		case REQUEST:
			if holder == -1 {
				holder = e.src
				e.resp <- Event{src: id, act: GRANT}
				//fmt.Printf("GRANT: fork %d to philo %d\n", id, e.src)
			} else {
				queue = append(queue, e)
			}

		case RELEASE:
			//fmt.Printf("RELEASE: fork %d by philo %d\n", id, e.src)
			holder = -1
		}
	}
}

func philosopher(id int, left, right chan Event, wg *sync.WaitGroup) {
	defer wg.Done() // wait group waits for this thread to finish

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
	meals := 0

	leftResp := make(chan Event)
	rightResp := make(chan Event)

	for meals < GOAL {
		fmt.Printf("THINKING: philo %d\n", id)
		time.Sleep(time.Millisecond * time.Duration(rng.Intn(10))) // prevents deadlock

		getFork(left, id, leftResp)
		getFork(right, (id+1)%N, rightResp)

		meals++
		fmt.Printf("EATING: philo %d (%d/%d)\n", id, meals, GOAL)

		left <- Event{src: id, act: RELEASE}
		right <- Event{src: id, act: RELEASE}
	}

	fmt.Printf("FINISHED: philo %d\n", id)
}

func getFork(fork chan Event, id int, resp chan Event) {
	fork <- Event{src: id, act: REQUEST, resp: resp}
	//fmt.Printf("REQUEST: philo %d for fork %d\n", id, id)
	<-resp
	//fmt.Printf("RECEIVE: philo %d fork %d\n", id, id)
}

func main() {
	var wg sync.WaitGroup
	chs := make([]chan Event, N)

	for i := 0; i < N; i++ {
		chs[i] = make(chan Event, 5)
		go fork(i, chs[i])
	}

	for i := 0; i < N; i++ {
		wg.Add(1)
		left := chs[i]
		right := chs[(i+1)%N]
		go philosopher(i, left, right, &wg)
	}

	wg.Wait() // waits for philo threads to return
	time.Sleep(10000)
	fmt.Printf("\n*** %d/%d philos ate %d times ***", N, N, GOAL)
}
