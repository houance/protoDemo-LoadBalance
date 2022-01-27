package main

import (
	"houance/protoDemo-LoadBalance/internal"
)

func main()  {
	internal.StartAllComponent("0.0.0.0:11000", "47.107.87.76:12000")
}