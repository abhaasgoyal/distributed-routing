# Introduction

The project consists of a distributed system of various topologies of routers which synchronously pass messages to each other in the shortest path. The next router to pass the message is decided using Distance Vector routing algorithm.

# Usage

`go run cmd/test_routers/main.go -t [Topology] -s [Size] -d [Dimensions] -w [Settling time] -m [Source/Destination Configuration] -x [Simulate Dropouts] -r [Number of repetitions]`

1. `Supported Topologies` - (a) by size: Line, Ring, Star, Fully_Connected
                              (b) by dimension and size : Mesh
                              (b) by dimension: Hypercube
2. `Source/Destination Configuration` - These are of 2 types - One_To_All (Message broadcast from a single random router) or All_To_One (All routers sending message to a single random router)
3. `Simulate Dropouts` (TODO) - In real world networks sometimes physical routers die out after connections have been established. After this, if there exists a path for the message to reach the destination one should find it. Implementation ideas would be to trap any failed messages in a go panic. Some techniques for these would be concurrent queues, poison reverse, etc.
4. `Number of repetitions` - Number of times the test should be repeated with a different random router in each iteration
5. `Force` - Force a large number of router creation (In case )

# Message structure

Various types of messages and their structures are decided before the routers are deployed and is encapsulated in Diagram.pdf
