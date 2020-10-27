package routers

import (
	"fmt"
	"sync"
)

type RouterId uint

type Template [][]RouterId

type Cost RouterId
type Envelope struct {
	Dest    RouterId
	Hops    uint
	Message interface{}
}

type RouterTable struct {
	// ID RouterId
	//table     []Destination // Array indices will tell us the neighbouring ID
	Next      []RouterId
	totalCost []uint
}

type TableMsg struct {
	LockRef *sync.WaitGroup
	rt      RouterTable
	Dest    []RouterId
}

type TestMessage int

func InitializeRouters(t Template) []RouterTable {
	// Here make the routing tables with lock step
	DistanceTable := make([]RouterTable, len(t))
	var wg sync.WaitGroup
	for routerId, neighbourIds := range t {
		wg.Add(1)
		go func(j RouterId, k []RouterId) {
			// ith router is DistanceTable[j]
			DistanceTable[j] = RouterTable{
				// ID : j
				Next:      make([]RouterId, len(t)),
				totalCost: make([]uint, len(t)),
			}
			// fmt.Println(k)
			for i := range DistanceTable[j].totalCost {
				DistanceTable[j].totalCost[i] = 1000000
			}
			for _, neighbourId := range k {
				// fmt.Println(neighbourId)
				DistanceTable[j].totalCost[neighbourId] = 1
				DistanceTable[j].Next[neighbourId] = RouterId(neighbourId)
			}
			//		fmt.Println(DistanceTable[j])
			wg.Done()
		}(RouterId(routerId), neighbourIds)
	}
	wg.Wait()
	//	fmt.Println("EVerything", DistanceTable[2])
	return DistanceTable
}

func MakeRouters(t Template) (in []chan<- interface{}, out <-chan Envelope) {
	channels := make([]chan interface{}, len(t))
	framework := make(chan Envelope)
	distanceFrame := make(chan TableMsg)

	in = make([]chan<- interface{}, len(t))
	for i := range channels {
		channels[i] = make(chan interface{})
		in[i] = channels[i]
	}
	out = framework
	tableList := InitializeRouters(t)
	fmt.Println(tableList[1])

	for routerId, neighbourIds := range t {
		// Make outgoing channels for each neighbours
		neighbours := make([]chan<- interface{}, len(neighbourIds))
		for i, id := range neighbourIds {
			neighbours[i] = channels[id]
		}
		// Decide on where to pass the message
		go Router(RouterId(routerId), channels[routerId], neighbours, framework, distanceFrame, &tableList[routerId])
	}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		for routerId, neighbourIds := range t {
			wg.Add(len(neighbourIds))
			//	Decide on where to pass the message
			go func(ID RouterId, neighbours []RouterId) {
				channels[ID] <- TableMsg{
					Dest:    neighbours,
					LockRef: &wg,
					rt: RouterTable{
						totalCost: tableList[ID].totalCost,
						Next:      tableList[ID].Next,
					}}
			}(RouterId(routerId), neighbourIds)

			//	<-channels[ID]
			//	fmt.Println("done")

			//	wg.Done()
		}
		wg.Wait()
	}

	return
}
