package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"routers"
)

var (
	topology = flag.String("t", "Line", "`topology` (by size: Line, Ring, Star, Fully_Connected; "+
		"by dimension and size: Mesh, Torus; by dimension: Hypercube, Cube_Connected_Cycles, Butterfly, Wrap_Around_Butterfly)")
	size             = flag.Uint("s", 5, "size")
	dimension        = flag.Uint("d", 2, "dimension")
	printConnections = flag.Bool("c", false, "print connections")
	printDistances   = flag.Bool("i", false, "print distances")
	settleTime       = flag.Duration("w", time.Second/10, "routers settle time")
	//timeout          = flag.Duration("o", time.Second/10, "comms timeout")
	mode     = flag.String("m", "One_To_All", "`mode` (One_To_All, All_To_One)")
	dropouts = flag.Uint("x", 0, "dropouts")
	repeats  = flag.Uint("r", 10, "repeats")
	force    = flag.Bool("f", false, "force the creation of a large number of routers")
)

func main() {
	flag.Parse()

	if *size == 0 && (*topology == "Line" || *topology == "Ring" || *topology == "Star" || *topology == "Fully_Connected" ||
		*topology == "Mesh" || *topology == "Torus") {
		fmt.Fprintln(os.Stderr, "You have requested a topology with zero routers. Try increasing size (-s).")
		os.Exit(1)
	}

	var template routers.Template

	switch *topology {
	case "Line", "Ring":
		template = make(routers.Template, *size)
		for i := routers.RouterId(0); uint(i) < *size; i++ {
			template[i] = []routers.RouterId{i - 1, i + 1}
		}
		if *topology == "Line" {
			template[0] = template[0][1:]
			template[*size-1] = template[*size-1][:len(template[*size-1])-1]
		} else {
			template[0][0] = routers.RouterId(*size - 1)
			template[*size-1][1] = 0
		}
	case "Star":
		template = make(routers.Template, *size)
		if *size > 0 {
			template[0] = make([]routers.RouterId, *size-1)
			for i := routers.RouterId(1); uint(i) < *size; i++ {
				template[0][i-1] = i
			}
			for i := uint(1); i < *size; i++ {
				template[i] = []routers.RouterId{0}
			}
		}
	case "Fully_Connected":
		template = make(routers.Template, *size)
		for i := routers.RouterId(0); uint(i) < *size; i++ {
			template[i] = make([]routers.RouterId, *size-1)
			for j := routers.RouterId(0); uint(j) < *size-1; j++ {
				if j < i {
					template[i][j] = j
				} else {
					template[i][j] = j + 1
				}
			}
		}
	case "Mesh":
		template = make(routers.Template, exp(*size, *dimension))
		for i := routers.RouterId(0); int(i) < len(template); i++ {
			temp := make(map[routers.RouterId]struct{})
			for d := uint(0); d < *dimension; d++ {
				if int(i)-(1<<d) >= 0 {
					temp[i-(1<<d)] = struct{}{}
				}
				if int(i)+(1<<d) < len(template) {
					temp[i+(1<<d)] = struct{}{}
				}
			}
			template[i] = make([]routers.RouterId, len(temp))
			j := 0
			for n := range temp {
				template[i][j] = n
				j++
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported topology %s\n", *topology)
		flag.Usage()
		os.Exit(1)
	}
	hops := make([][]uint, len(template))
	for hopI := range hops {
		hops[hopI] = make([]uint, len(template))
	}

	var maxHops uint = 0
	var minHops = ^uint(0)
	var totalMessages float64 = 0
	var sumHops float64 = 0

	in, out := routers.MakeRouters(template)
	if *printConnections {
		PrintConnections(template)
	}
	time.Sleep(*settleTime)
	start := time.Now()

	msgs := make(map[uint]struct{})

	for i := routers.RouterId(0); int(i) < len(template); i++ {
		msgs[uint(i)] = struct{}{}
		go func(j routers.RouterId) {
			switch *mode {
			// Bug or feature of sending it to itself ? Handled in receiving tho
			case "One_To_All":
				in[0] <- routers.Envelope{
					Dest:    j,
					Hops:    0,
					Message: uint(j),
				}
			case "All_To_One":
				in[j] <- routers.Envelope{
					Dest:    0,
					Hops:    0,
					Message: uint(j),
				}
			default:
				fmt.Fprintf(os.Stderr, "Unsupported test mode %s\n", *mode)
				flag.Usage()
				os.Exit(1)
			}
		}(i)
	}

	for {
		envelope := <-out
		if i, ok := envelope.Message.(uint); ok {
			if _, ok := msgs[i]; ok {
				hops[envelope.Message.(uint)][envelope.Dest] = envelope.Hops
				if envelope.Hops != 0 {
					sumHops += float64(envelope.Hops)
					totalMessages += 1.0
					if envelope.Hops > maxHops {
						maxHops = envelope.Hops
					}
					if envelope.Hops < minHops {
						minHops = envelope.Hops
					}
				}
				delete(msgs, i)
				if len(msgs) == 0 {
					break
				}
			} else {
				log.Printf("Unexpected message value %v! Make sure you aren't duplicating Envelopes.", i)
			}
		} else {
			log.Printf("Unexpected message body %g! Make sure you aren't editing Envelopes.", envelope.Message)
		}
	}

	fmt.Printf("\n------------------------------------------------------------------------------\n")
	log.Printf("Test completed in %v\n", time.Since(start))

	fmt.Printf("------------------------------ Measurements ------------------------------------\n")
	fmt.Printf("Average Hops: %.2f \n", sumHops/totalMessages)
	fmt.Printf("Minimum Hops: %d\n", minHops)
	fmt.Printf("Maximum Hops: %d\n", maxHops)
	if *printDistances {
		PrintDistances(template, hops)
	}
}

func exp(x uint, y uint) uint {
	z := new(big.Int).Exp(new(big.Int).SetUint64(uint64(*size)), new(big.Int).SetUint64(uint64(*dimension)), nil)
	if z.Cmp(big.NewInt(1024)) > 0 && !*force {
		fmt.Fprintf(os.Stderr, "Your chosen configuration would generate a very large number of routers.\n"+
			"Try setting dimension (-d) smaller. (Would generate %v routers.)\n"+
			"If you're really sure, you can use -f to proceed anyway.\n", z)
		os.Exit(1)
	}
	if !z.IsUint64() {
		fmt.Fprintln(os.Stderr, "OK, but seriously though, this would generate more than 2^64 routers. Aborting.")
		os.Exit(1)
	}
	return uint(z.Uint64())
}

// Below functions algorithm and design are inspired from the Ada version assignment
func PrintConnections(t [][]routers.RouterId) {
	fmt.Print("\n")
	fmt.Println("--------------- Information about the selected network topology ----------------")

	fmt.Print("\n")
	fmt.Print("    ")
	for i := range t {
		fmt.Printf("%3d", i)
	}
	fmt.Print("\n")
	fmt.Print("    +")
	fmt.Print(strings.Repeat("---", len(t)))
	fmt.Print("+\n")
	for i := range t {
		fmt.Printf("%3d", i)
		fmt.Print(" |")
		for j := range t {
			if i == j {
				fmt.Print(" . ")
			} else if NodesConnected(routers.RouterId(j), t[i]) {
				fmt.Print("<->")
			} else {
				fmt.Print("   ")
			}
		}
		fmt.Print("|\n")
	}
	fmt.Print("    +")
	fmt.Print(strings.Repeat("---", len(t)))
	fmt.Print("+\n")

}

func PrintDistances(t [][]routers.RouterId, hops [][]uint) {
	fmt.Print("\n")
	fmt.Print("    ")
	for i := range t {
		fmt.Printf("%3d", i)
	}
	fmt.Print("\n")
	fmt.Print("    +")
	fmt.Print(strings.Repeat("---", len(t)))
	fmt.Print("+\n")
	for i := range t {
		fmt.Printf("%3d", i)
		fmt.Print(" |")
		for j := range t {
			if hops[i][j] != 0 {
				fmt.Printf("%3d", hops[i][j])
			} else if hops[j][i] != 0 {
				fmt.Printf("%3d", hops[j][i])
			} else {
				fmt.Print("   ")
			}
		}
		fmt.Print("|\n")
	}
	fmt.Print("    +")
	fmt.Print(strings.Repeat("---", len(t)))
	fmt.Print("+\n")

}

func NodesConnected(node routers.RouterId, t []routers.RouterId) bool {
	for _, neighbour := range t {
		if neighbour == node {
			return true
		}
	}
	return false
}
