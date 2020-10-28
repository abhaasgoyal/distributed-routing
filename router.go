package routers

import (
	"log"
)

func Router(self RouterId, incoming <-chan interface{}, neighbours []chan<- interface{}, framework chan<- Envelope, neighbourIds []RouterId, table *RouterTable) {
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
						if id == table.Next[msg.Dest] {
							neighbours[i] <- msg
						}
					}
				}
			// Add more cases to handle any other message types you create here
			case TableMsg:
				if ReceivingEnd(self, msg.Dest) {
					// Receive and Update the routerTable
					for id, Cost := range msg.Costs {
						if table.Costs[id] > Cost+1 {
							table.Next[id] = msg.Sender
							table.Costs[id] = Cost + 1
						}
					}
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
