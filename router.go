package routers

import (
	"log"
)

func Router(self RouterId, incoming <-chan interface{}, neighbours []chan<- interface{}, framework chan<- Envelope, distanceFrame chan<- TableMsg, table *RouterTable) {
	for {
		select {
		// The operation blocks here before the time limit I guess in the case of envelope
		case raw := <-incoming:
			switch msg := raw.(type) {
			case Envelope:
				if msg.Dest == self {
					framework <- msg
				} else {
					// Handle forwarding on a message here
					msg.Hops += 1
					neighbours[0] <- msg
				}
			// Add more cases to handle any other message types you create here
			case TableMsg:
				if ReceivingEnd(self, msg.Dest) {
					// fmt.Println(msg.rt)
					// Update the routerTable
					for id, Cost := range msg.Costs {
						// if Cost+1-table.Costs[id] > 1 {
						//	fmt.Println(table.Costs[id], Cost+1)
						// }
						if table.Costs[id] > Cost+1 {
							table.Next[id] = msg.Sender
							table.Costs[id] = Cost + 1
						}
					}
					msg.LockRef.Done()
				} else {
					// Handle forwarding on a message here
					for i := range neighbours {
						go func(j int) {
							neighbours[j] <- msg
						}(i)
					}
				}
				// distanceFrame <- msg
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
