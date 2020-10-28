package routers

import (
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

// RouterTable of single router
type RouterTable struct {
	Next  []RouterId
	Costs []uint
}

// Messages to pass to implement DVR protocol
type TableMsg struct {
	Costs  []uint
	Dest   []RouterId
	Sender RouterId
}

type TestMessage int

func InitializeRoutingTable(t Template) []RouterTable {
	// Here make the routing tables with lock step
	DistanceTable := make([]RouterTable, len(t))
	const infinity = 1000000
	var routerGroup sync.WaitGroup
	for routerId, neighbourIds := range t {
		routerGroup.Add(1)
		go func(self RouterId, neighbours []RouterId) {
			DistanceTable[self] = RouterTable{
				Next:  make([]RouterId, len(t)),
				Costs: make([]uint, len(t)),
			}
			var initialUpdateGroup sync.WaitGroup
			for i := range DistanceTable[self].Costs {
				initialUpdateGroup.Add(1)
				go func(CostIdx int) {
					defer initialUpdateGroup.Done()
					DistanceTable[self].Costs[CostIdx] = infinity
				}(i)
			}
			initialUpdateGroup.Wait()
			DistanceTable[self].Costs[self] = 0

			for _, neighbourId := range neighbours {
				initialUpdateGroup.Add(1)
				go func(nIdx RouterId) {
					defer initialUpdateGroup.Done()
					DistanceTable[self].Costs[nIdx] = 1
					DistanceTable[self].Next[nIdx] = nIdx
				}(neighbourId)
			}
			initialUpdateGroup.Wait()
			routerGroup.Done()
		}(RouterId(routerId), neighbourIds)
	}
	routerGroup.Wait()
	return DistanceTable
}

func MakeRouters(t Template) (in []chan<- interface{}, out <-chan Envelope) {

	// Array of channels
	channels := make([]chan interface{}, len(t))
	framework := make(chan Envelope)
	// distanceFrame := make(chan TableMsg)

	in = make([]chan<- interface{}, len(t))
	for i := range channels {
		// Synchronized channel for each router
		channels[i] = make(chan interface{})
		in[i] = channels[i]
	}
	out = framework

	// tableList is the list for all routing tables
	// Specific routing table reference will be sent
	tableList := InitializeRoutingTable(t)

	for routerId, neighbourIds := range t {
		// Make outgoing channels for each neighbours
		neighbours := make([]chan<- interface{}, len(neighbourIds))
		for i, id := range neighbourIds {
			neighbours[i] = channels[id]
		}
		// Decide on where to pass the message
		// We send two additional parameters - reference to the router's own table and the ID's of it's neighbours (not the channel)
		go Router(RouterId(routerId), channels[routerId], neighbours, framework, neighbourIds, &tableList[routerId])
	}

	// Create a copy of the cost table for the iteration
	// Because slices' values are passed by reference
	costCopy := make([][]uint, len(t))
	for i := range tableList {
		costCopy[i] = make([]uint, len(t))
		copy(costCopy[i], tableList[i].Costs)
	}

	for routerId, neighbourIds := range t {
		//	Decide on where to pass the message for routing tables
		go func(ID RouterId, neighbours []RouterId) {
			channels[ID] <- TableMsg{
				Dest:   neighbours,
				Costs:  costCopy[ID],
				Sender: ID,
			}
		}(RouterId(routerId), neighbourIds)
	}

	return
}
