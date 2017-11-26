package main

import (
	"log"
	"net"
	"bufio"
	"regexp"
	"strings"
)

type IServer interface {
	Listen(port string) error
	ParseMessage()
}

type Server struct {
	Listener net.Listener
	Port string
	Sem chan int
	Matcher *regexp.Regexp
	Shutdown chan bool
	Killed bool
}

func NewServer(port string, maxConnections int) *Server {
	regex, _ := regexp.Compile(`^0*[1-9]\d{6,9}$`)

	return &Server {
		Port: port,
		Sem: make(chan int, maxConnections),
		Shutdown: make(chan bool),
		Matcher: regex,
	}
}

func (s *Server) Listen() {
	var err error
	if s.Listener, err = net.Listen("tcp", s.Port); err != nil {
		log.Printf("Failed to create TCP server listening on port %s\nErr: %s\n", s.Port, err.Error())
		return
	}

	for {
		select {
		case <-s.Shutdown:
			log.Println("Received shutdown command on", s.Listener.Addr())
			s.Listener.Close()
			return
		default:
			s.Sem <- 1
		}
		
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection\n%s\n", err.Error())
			continue
		}
		log.Printf("Accepted connection\n")
		
		go s.ParseMessage(conn)
	}
}

func (s *Server) ParseMessage(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Closing connection", err)
			<- s.Sem
			return
		}

		input = strings.TrimRight(input, "\r\n")

		if input == "shutdown" {
			s.Shutdown <- true
			conn.Close()
			return
		} else if !s.Matcher.MatchString(input) {
			log.Println("Received malformed data, closing connection")
			conn.Close()
			<- s.Sem
			return
		}

		log.Printf("%s", input)
	}
}
