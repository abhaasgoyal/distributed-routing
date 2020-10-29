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
					// Send the real world message to the routing table
					// specification
					msg.Hops += 1
					for i, id := range neighbourIds {
						if id == localTable.Next[msg.Dest] {
							neighbours[i] <- msg
						}
					}
				}
			case TableMsg:
				if ReceivingEnd(self, msg.Dest) {
					// Receive and Update the routerTable
					sameTable := true
					for id, Cost := range msg.Costs {
						if Cost+1 < localTable.Costs[id] {
							sameTable = false
							localTable.Next[id] = msg.Sender
							localTable.Costs[id] = Cost + 1
						}
					}
					// Repeat until convergence (setup time: diameter)
					if !sameTable {
						costCopy := make([]uint, len(localTable.Costs))
						copy(costCopy, localTable.Costs)
						for i := range neighbours {
							go func(j int) {
								neighbours[j] <- TableMsg{
									Dest:   neighbourIds,
									Costs:  costCopy,
									Sender: self,
								}
							}(i)
						}
					}
				} else {
					// Handle forwarding of Routing Table
					for i := range neighbours {
						go func(j int) {
							neighbours[j] <- TableMsg{
								Dest:   neighbourIds,
								Costs:  msg.Costs,
								Sender: self,
							}
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
