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

type TableMessage struct {
	operation string
	table     RouterTable
}

type RouterTable struct {
	// ID RouterId
	table []Destination // Array indices will tell us the neighbouring ID
}
type Destination struct {
	Next      RouterId
	totalCost uint
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
				table: make([]Destination, len(t)),
			}
			// fmt.Println(k)
			for i := range DistanceTable[j].table {
				DistanceTable[j].table[i].totalCost = 1000000

			}
			for _, neighbourId := range k {
				// fmt.Println(neighbourId)
				DistanceTable[j].table[neighbourId].totalCost = 1
				DistanceTable[j].table[neighbourId].Next = RouterId(neighbourId)
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

	in = make([]chan<- interface{}, len(t))
	for i := range channels {
		channels[i] = make(chan interface{})
		in[i] = channels[i]
	}
	out = framework
	tableList := InitializeRouters(t)
	fmt.Println(tableList[1])

	var wg sync.WaitGroup
	for routerId, neighbourIds := range t {
		wg.Add(2)
		// Make outgoing channels for each neighbours
		neighbours := make([]chan<- interface{}, len(neighbourIds))
		for i, id := range neighbourIds {
			neighbours[i] = channels[id]
		}
		// Decide on where to pass the message
		go Router(RouterId(routerId), channels[routerId], neighbours, framework, &tableList[routerId])
		go func(ID int) {
			channels[ID] <- RouterTable{
				table: tableList[ID].table,
			}
			wg.Done()
		}(routerId)
		// go func(ID int) {
		//	<-channels[ID]
		//	fmt.Println("done")

		//	wg.Done()
		// }(routerId)
	}
	wg.Wait()

	for routerId, neighbourIds := range t {
		// Make outgoing channels for each neighbours
		neighbours := make([]chan<- interface{}, len(neighbourIds))
		for i, id := range neighbourIds {
			neighbours[i] = channels[id]
		}
		// Decide on where to pass the message
		go Router(RouterId(routerId), channels[routerId], neighbours, framework, &tableList[routerId])
	}

	return
}
