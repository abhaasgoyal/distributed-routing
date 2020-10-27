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

type RouterTable struct {
	// ID RouterId
	//table     []Destination // Array indices will tell us the neighbouring ID
	Next  []RouterId
	Costs []uint
}

type TableMsg struct {
	LockRef *sync.WaitGroup
	Costs   []uint
	Dest    []RouterId
	Sender  RouterId
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
				Next:  make([]RouterId, len(t)),
				Costs: make([]uint, len(t)),
			}
			// fmt.Println(k)
			for i := range DistanceTable[j].Costs {
				DistanceTable[j].Costs[i] = 1000000
				//				DistanceTable[j].Next[i] = -1
			}
			DistanceTable[j].Costs[j] = 0
			for _, neighbourId := range k {
				// fmt.Println(neighbourId)
				DistanceTable[j].Costs[neighbourId] = 1
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
	// distanceFrame := make(chan TableMsg)

	in = make([]chan<- interface{}, len(t))
	for i := range channels {
		channels[i] = make(chan interface{})
		in[i] = channels[i]
	}
	out = framework
	tableList := InitializeRouters(t)
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
	for i := 0; i < 1000; i++ {
		for routerId, neighbourIds := range t {
			wg.Add(len(neighbourIds))

			//	Decide on where to pass the message
			go func(ID RouterId, neighbours []RouterId) {
				channels[ID] <- TableMsg{
					Dest:    neighbours,
					LockRef: &wg,
					Costs:   tableList[ID].Costs,
					Sender:  ID,
					// rt: RouterTable{
					//	totalCost: tableList[ID].totalCost,
					//	Next:      tableList[ID].Next,
					// }
				}
				//				fmt.Println(ID)
			}(RouterId(routerId), neighbourIds)

			//	<-channels[ID]
			//	fmt.Println("done")

			//	wg.Done()
		}
		wg.Wait()
	}
	// fmt.Println(tableList[0])

	return
}
