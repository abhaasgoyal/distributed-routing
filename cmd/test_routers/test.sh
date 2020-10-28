#!/bin/bash

# 2 dimension basics all to one
go run -race main.go -t Line -s 20 -d 2 -m All_To_One -x 0 -r 10
go run -race main.go -t Line -s 600 -d 2 -m All_To_One -x 0 -r 10
go run -race main.go -t Ring -s 20 -d 2 -m All_To_One -x 0 -r 10
go run -race main.go -t Ring -s 600 -d 2 -m All_To_One -x 0 -r 10
go run -race main.go -t Star -s 20 -d 2 -m All_To_One -x 0 -r 10
go run -race main.go -t Star -s 600 -d 2 -m All_To_One -x 0 -r 10

# 3 dimension basics all to one
go run -race main.go -t Line -s 20 -d 3 -m All_To_One -x 0 -r 10
go run -race main.go -t Line -s 600 -d 3 -m All_To_One -x 0 -r 10
go run -race main.go -t Ring -s 20 -d 3 -m All_To_One -x 0 -r 10
go run -race main.go -t Ring -s 600 -d 3 -m All_To_One -x 0 -r 10
go run -race main.go -t Star -s 20 -d 3 -m All_To_One -x 0 -r 10
go run -race main.go -t Star -s 600 -d 3 -m All_To_One -x 0 -r 10

# 2 dimension complex all to one
go run -race main.go -t Fully_Connected -s 20 -d 2 -m All_To_One -x 0 -r 10
go run -race main.go -t Mesh -s 10 -d 3 -m All_To_One -x 0 -r 10


# 2 dimension basics  one to all
go run -race main.go -t Line -s 20 -d 2 -m One_To_All -x 0 -r 10
go run -race main.go -t Line -s 600 -d 2 -m One_To_All -x 0 -r 10
go run -race main.go -t Ring -s 20 -d 2 -m One_To_All -x 0 -r 10
go run -race main.go -t Ring -s 600 -d 2 -m One_To_All -x 0 -r 10
go run -race main.go -t Star -s 20 -d 2 -m One_To_All -x 0 -r 10

# 3 dimension basics  one to all
go run -race main.go -t Star -s 600 -d 3 -m One_To_All -x 0 -r 10
go run -race main.go -t Line -s 600 -d 3 -m One_To_All -x 0 -r 10
go run -race main.go -t Ring -s 20 -d 3 -m One_To_All -x 0 -r 10
go run -race main.go -t Ring -s 600 -d 3 -m One_To_All -x 0 -r 10
go run -race main.go -t Star -s 20 -d 3 -m One_To_All -x 0 -r 10
go run -race main.go -t Star -s 600 -d 3 -m One_To_All -x 0 -r 10

# 2 dimension complex  one to all
go run -race main.go -t Fully_Connected -s 20 -d 2 -m One_To_All -x 0 -r 10
go run -race main.go -t Mesh -s 20 -d 2 -m One_To_All -x 0 -r 10
