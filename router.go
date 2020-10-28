package routers

import (
	"log"
)

func Router(self RouterId, incoming <-chan interface{}, neighbours []chan<- interface{}, framework chan<- Envelope, neighbourIds []RouterId, localTable *RouterTable) {
	for {
		select {
		// The operation blocks here and repeatedly listens for calls
		case raw := <-incoming:
			switch msg := raw.(type) {
			case Envelope:
				if msg.Dest == self {
					framework <- msg
				} else {
					// Send the real world message to the routing table specs
					msg.Hops += 1
					for i, id := range neighbourIds {
						if id == localTable.Next[msg.Dest] {
							neighbours[i] <- msg
						}
					}
				}
			// Add more cases to handle any other message types you create here
			case TableMsg:
				if ReceivingEnd(self, msg.Dest) {
					// Receive and Update the routerTable
					// var routeGroup sync.WaitGroup
					for id, Cost := range msg.Costs {
						// routeGroup.Add(1)
						// go func(ID int, nCost uint) {
						//	defer routeGroup.Done()
						if Cost+1 < localTable.Costs[id] {
							localTable.Next[id] = msg.Sender
							localTable.Costs[id] = Cost + 1
						}
						// }(id, Cost)
					}
					// routeGroup.Wait()
					msg.LockRef.Done()
				} else {
					// Handle forwarding of Routing Table
					for i := range neighbours {
						go func(j int) {
							neighbours[j] <- msg
						}(i)
					}
				}
			default:
				log.Printf("[%v] received unexpected message %g\n", self, msg)
			}
		}
	}
}

func ReceivingEnd(item RouterId, DestIDs []RouterId) bool {
	for i := 0; i < len(DestIDs); i++ {
		if DestIDs[i] == item {
			return true
		}
	}
	return false
}
