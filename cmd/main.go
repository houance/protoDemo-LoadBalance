package main

import (
	"houance/protoDemo-LoadBalance/internal"
)

func main()  {
	internal.StartAllComponent("127.0.0.1:11000", "127.0.0.1:12000")
}