package routers

import (
	"log"
)

func Router(self RouterId, incoming <-chan interface{}, neighbours []chan<- interface{}, framework chan<- Envelope, neighbourIds []RouterId, table *RouterTable) {
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
					for i, id := range neighbourIds {
						//	fmt.Println(self)
						//fmt.Println(i, id, table.Next[self])
						if id == table.Next[msg.Dest] {
							neighbours[i] <- msg
						}
					}
					//	neighbours[0] <- msg
				}
			// Add more cases to handle any other message types you create here
			case TableMsg:
				if ReceivingEnd(self, msg.Dest) {
					// fmt.Println(msg.rt)
					// Update the routerTable
					// var updateRec sync.WaitGroup
					for id, Cost := range msg.Costs {
						// updateRec.Add(1)
						// go func(ID int, sendCost uint) {
						//	defer updateRec.Done()
						if table.Costs[id] > Cost+1 {
							table.Next[id] = msg.Sender
							table.Costs[id] = Cost + 1
						}
						// }(id, Cost)
					}
					// updateRec.Wait()
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
