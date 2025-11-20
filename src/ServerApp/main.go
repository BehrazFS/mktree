package main

import server "ServerApp/Server"

func main() {
	server := server.NewServer()
	server.StartServer("8080")
}
