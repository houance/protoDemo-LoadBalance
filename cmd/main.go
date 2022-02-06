package main

import (
	"houance/protoDemo-LoadBalance/internal"
)

func main()  {
	internal.StartAllComponent("0.0.0.0:11000", "172.30.58.167:12000")
}