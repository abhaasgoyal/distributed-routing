package routers

import (
	"fmt"
	"log"
)

func Router(self RouterId, incoming <-chan interface{}, neighbours []chan<- interface{}, framework chan<- Envelope, table *RouterTable) {
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
			case RouterTable:
				fmt.Println("Woah")
			default:
				log.Printf("[%v] received unexpected message %g\n", self, msg)
			}
		}
	}
}
