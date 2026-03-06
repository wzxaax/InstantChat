package main

import "InstantChat/kernal"

func main() {
	Server := kernal.NewServer("127.0.0.1", 8888)
	Server.Start()
}
