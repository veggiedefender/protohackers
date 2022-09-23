package budgetchat

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
)

type Name string

type Server struct {
	Clients    map[Name]*Client
	ClientsMux sync.RWMutex
}

type Client struct {
	Name       Name
	Inbox      chan string
	Outbox     chan string
	Disconnect chan interface{}
}

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func NewServer() *Server {
	return &Server{
		Clients:    make(map[Name]*Client),
		ClientsMux: sync.RWMutex{},
	}
}

func (s *Server) Listen(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("tcp server accept error", err)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	_, err := conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))
	if err != nil {
		log.Println(err)
		return
	}
	if !scanner.Scan() {
		return
	}
	name, err := s.validateName(scanner.Text())
	if err != nil {
		return
	}

	client := Client{
		Name:       name,
		Inbox:      make(chan string, 10),
		Outbox:     make(chan string, 1),
		Disconnect: make(chan interface{}),
	}
	s.registerClient(&client)
	defer s.disconnectClient(&client)

	go client.readInputs(scanner)

	for {
		select {
		case <-client.Disconnect:
			return
		case msg := <-client.Inbox:
			_, err := conn.Write([]byte(msg + "\n"))
			if err != nil {
				return
			}
		case msg := <-client.Outbox:
			s.broadcast(&client, msg)
		}
	}
}

func (s *Server) registerClient(client *Client) {
	s.broadcastAll(fmt.Sprintf("%s has entered the room", client.Name))
	client.Inbox <- fmt.Sprintf("* The room contains: %s", strings.Join(s.listClientNames(), ", "))

	s.ClientsMux.Lock()
	s.Clients[client.Name] = client
	s.ClientsMux.Unlock()
}

func (s *Server) disconnectClient(client *Client) {
	s.ClientsMux.Lock()
	delete(s.Clients, client.Name)
	s.ClientsMux.Unlock()

	s.broadcastAll(fmt.Sprintf("%s has left the room", client.Name))
}

func (s *Server) broadcast(sender *Client, msg string) {
	s.ClientsMux.RLock()
	defer s.ClientsMux.RUnlock()

	msg = fmt.Sprintf("[%s] %s", sender.Name, msg)

	for _, recipient := range s.Clients {
		if recipient.Name == sender.Name {
			continue
		}

		recipient.Inbox <- msg
	}
}

func (s *Server) broadcastAll(msg string) {
	s.ClientsMux.RLock()
	defer s.ClientsMux.RUnlock()

	msg = fmt.Sprintf("* %s", msg)

	for _, recipient := range s.Clients {
		recipient.Inbox <- msg
	}
}

func (s *Server) listClientNames() []string {
	s.ClientsMux.RLock()
	defer s.ClientsMux.RUnlock()

	names := make([]string, 0, len(s.Clients))
	for _, client := range s.Clients {
		names = append(names, string(client.Name))
	}
	return names
}

func (c *Client) readInputs(scanner *bufio.Scanner) {
	for scanner.Scan() {
		c.Outbox <- scanner.Text()
	}
	close(c.Disconnect)
}

func (s *Server) validateName(name string) (Name, error) {
	if !nameRegex.Match([]byte(name)) {
		return "", errors.New("invalid name")
	}

	s.ClientsMux.RLock()
	defer s.ClientsMux.RUnlock()

	if _, ok := s.Clients[Name(name)]; ok {
		return "", errors.New("name already used")
	}

	return Name(name), nil
}
