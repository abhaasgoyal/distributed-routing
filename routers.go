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

// RouterTable of single router (identified by ID)
type RouterTable struct {
	Next  []RouterId
	Costs []uint
}

// Messages to pass to implement DVR protocol
type TableMsg struct {
	LockRef *sync.WaitGroup
	Costs   []uint
	Dest    []RouterId
	Sender  RouterId
}

type TestMessage int

func InitializeRoutingTable(t Template) []RouterTable {
	// Here make the routing tables with lock step
	DistanceTable := make([]RouterTable, len(t))
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
					DistanceTable[self].Costs[CostIdx] = 1000000
				}(i)
			}
			initialUpdateGroup.Wait()
			DistanceTable[self].Costs[self] = 0

			for _, neighbourId := range neighbours {
				initialUpdateGroup.Add(1)
				go func(nIdx RouterId) {
					defer initialUpdateGroup.Done()
					DistanceTable[self].Costs[nIdx] = 1
					DistanceTable[self].Next[nIdx] = RouterId(nIdx)
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
	channels := make([]chan interface{}, len(t))
	framework := make(chan Envelope)
	// distanceFrame := make(chan TableMsg)

	in = make([]chan<- interface{}, len(t))
	for i := range channels {
		channels[i] = make(chan interface{})
		in[i] = channels[i]
	}
	out = framework

	// tableList is the list for all tables
	tableList := InitializeRoutingTable(t)

	for routerId, neighbourIds := range t {
		// Make outgoing channels for each neighbours
		neighbours := make([]chan<- interface{}, len(neighbourIds))
		for i, id := range neighbourIds {
			neighbours[i] = channels[id]
		}
		// Decide on where to pass the message
		// We send two additional parameters - reference to the router's own table and the ID's of it's neighbours
		go Router(RouterId(routerId), channels[routerId], neighbours, framework, neighbourIds, &tableList[routerId])
	}

	var updateGroup sync.WaitGroup
	for i := 0; i < len(t); i++ {
		for routerId, neighbourIds := range t {
			updateGroup.Add(len(neighbourIds))
			//	Decide on where to pass the message for routing tables
			go func(ID RouterId, neighbours []RouterId) {
				channels[ID] <- TableMsg{
					Dest:    neighbours,
					LockRef: &updateGroup,
					Costs:   tableList[ID].Costs,
					Sender:  ID,
				}
			}(RouterId(routerId), neighbourIds)
		}
		updateGroup.Wait()
	}

	return
}
