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
	var wg sync.WaitGroup
	for routerId, neighbourIds := range t {
		wg.Add(1)
		go func(j RouterId, k []RouterId) {
			// ith router is DistanceTable[j]
			DistanceTable[j] = RouterTable{
				// ID : j
				Next:  make([]RouterId, len(t)),
				Costs: make([]uint, len(t)),
			}
			var updateDis sync.WaitGroup
			for i := range DistanceTable[j].Costs {
				updateDis.Add(1)
				go func(CostIdx int) {
					defer updateDis.Done()
					DistanceTable[j].Costs[CostIdx] = 1000000
				}(i)
			}
			updateDis.Wait()
			DistanceTable[j].Costs[j] = 0

			for _, neighbourId := range k {
				updateDis.Add(1)
				// fmt.Println(neighbourId)
				go func(ID RouterId) {
					defer updateDis.Done()
					DistanceTable[j].Costs[ID] = 1
					DistanceTable[j].Next[ID] = RouterId(ID)
				}(neighbourId)
			}
			updateDis.Wait()
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
	// distanceFrame := make(chan TableMsg)

	in = make([]chan<- interface{}, len(t))
	for i := range channels {
		channels[i] = make(chan interface{})
		in[i] = channels[i]
	}
	out = framework
	tableList := InitializeRoutingTable(t)
	// fmt.Println(tableList[1])

	for routerId, neighbourIds := range t {
		// Make outgoing channels for each neighbours
		neighbours := make([]chan<- interface{}, len(neighbourIds))
		for i, id := range neighbourIds {
			neighbours[i] = channels[id]
			//fmt.Println(i, id)
		}
		// Decide on where to pass the message
		go Router(RouterId(routerId), channels[routerId], neighbours, framework, neighbourIds, &tableList[routerId])
	}

	var wg sync.WaitGroup
	for i := 0; i < len(t); i++ {
		for routerId, neighbourIds := range t {
			wg.Add(len(neighbourIds))

			//	Decide on where to pass the message
			go func(ID RouterId, neighbours []RouterId) {
				channels[ID] <- TableMsg{
					Dest:    neighbours,
					LockRef: &wg,
					Costs:   tableList[ID].Costs,
					Sender:  ID,
				}
				//				fmt.Println(ID)
			}(RouterId(routerId), neighbourIds)
		}
		wg.Wait()
	}
	// fmt.Println(tableList[0])

	return
}
