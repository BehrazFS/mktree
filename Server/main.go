package main

import server "mktree/Server"

func main() {
	server := server.NewServer()
	server.StartServer("8080")
}
