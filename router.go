package routers

import (
	"log"
	"sync"
	"time"
)

type some struct {
	Mutex sync.RWMutex
}

func Router(self RouterId, incoming <-chan interface{}, neighbours []chan<- interface{}, framework chan<- Envelope, neighbourIds []RouterId, localTable *RouterTable) {

	// Incase the path that we go
	queue := make([]Envelope, 0)
	// For concurrent entries within a router
	var queueLock sync.RWMutex

loop:
	for {
		select {
		// The operation blocks here and repeatedly listens for calls
		case raw := <-incoming:
			switch msg := raw.(type) {
			case Envelope:
				defer func() {
					if r := recover(); r != nil {
						var nID RouterId
						for i, id := range neighbourIds {
							if id == localTable.Next[msg.Dest] {
								// fmt.Println(self, id, msg.Dest)
								nID = RouterId(i)
								break
							}
						}
						localTable.Next[msg.Dest] = infinity
						neighbours = append(neighbours[:nID], neighbours[nID+1:]...)
						neighbourIds = append(neighbourIds[:nID], neighbourIds[nID+1:]...)
						log.Printf("Woah!!!! Handle Panic after sending")
						// framework <- msg
						go func() {
							var wt sync.WaitGroup
							wt.Add(1)
							queueLock.Lock()
							queue = append(queue, msg)
							queueLock.Unlock()
							PassOn(SecretPass{
								DeadGroup: &wt,
								Cost:      infinity,
								DeadID:    localTable.Next[msg.Dest],
								Neighbour: self,
								Dest:      msg.Dest,
								Sender:    self,
							}, neighbours)
							wt.Wait()
							queueLock.Lock()
							for i, id := range neighbourIds {
								if id == localTable.Next[msg.Dest] {
									neighbours[i] <- queue[0]
									break
								}
							}
							queue = queue[1:]
							queueLock.Unlock()
						}()
					}
				}()

				if msg.Dest == self {
					framework <- msg
				} else {
					// Send the real world message to the routing table
					// specification
					msg.Hops += 1
					for i, id := range neighbourIds {
						if id == localTable.Next[msg.Dest] {
							neighbours[i] <- msg
							break
						}
					}
				}

			case TableMsg:
				if IsNeighbour(self, msg.Dest) {
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
						PassOn(TableMsg{
							Dest:   neighbourIds,
							Costs:  costCopy,
							Sender: self,
						}, neighbours)
					}
				} else {
					// Handle forwarding of Routing Table in first run
					PassOn(TableMsg{
						Dest:   neighbourIds,
						Costs:  msg.Costs,
						Sender: self,
					}, neighbours)
				}
			case Death:
				for i := range neighbours {
					close(neighbours[i])
				}
				break loop
			case SecretPass:
				var tempCost uint
				var sameVal bool = true

				// If received back a new route send from there
				if self == msg.Neighbour {
					msg.DeadGroup.Done()
					break
				}

				// If received a new type of msg cost change it
				if msg.Cost != localTable.Costs[msg.DeadID] {
					if msg.Cost+1 < localTable.Costs[msg.Dest] {
						localTable.Costs[msg.DeadID] = msg.Cost
						localTable.Next[msg.Dest] = msg.Sender
					} else if msg.Cost == infinity {
						localTable.Costs[msg.DeadID] = infinity
						localTable.Next[msg.Dest] = msg.Sender
					}
					sameVal = false
				}

				// Keep propagating infinity till a neighbour is found
				if msg.Cost == infinity {
					if IsNeighbour(msg.Dest, neighbourIds) {
						tempCost = 1
					} else {
						tempCost = infinity
					}
				} else {
					tempCost = msg.Cost
				}
				if !sameVal {
					PassOn(SecretPass{
						Cost:      tempCost,
						DeadID:    msg.DeadID,
						DeadGroup: msg.DeadGroup,
						Neighbour: msg.Neighbour,
						Dest:      msg.Dest,
						Sender:    self,
					}, neighbours)
				}
			case nil:
				// This is happening sometimes on panic
			default:
				log.Printf("[%v] received unexpected message %g\n", self, msg)
			}
		}
	}
}

func IsNeighbour(source RouterId, neighbours []RouterId) bool {
	for _, id := range neighbours {
		if id == source {
			return true
		}
	}
	return false
}

func PassOn(message interface{}, neighbours []chan<- interface{}) {
	defer func() {
		if r := recover(); r != nil {
			time.Sleep(20000)
			log.Print(neighbours, message)
		}
	}()
	for i := range neighbours {
		go func(j int) {
			neighbours[j] <- message
		}(i)
	}
}
