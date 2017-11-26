package main

func main() {
	server := NewServer(":3280", 6)
	server.Listen()
}
