package speeddaemon

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct{}

type Client struct {
	HasSentWantHeartbeat bool
	Writer               *bufio.Writer
	WriterMu             sync.Mutex

	Camera     *MessageIAmCamera
	Dispatcher *MessageIAmDispatcher

	Closed chan interface{}
}

const Decisecond = time.Second / 10

func handleConnection(ticketer *Ticketer, dispatcherTracker *DispatcherTracker, conn net.Conn) error {
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	client := &Client{
		HasSentWantHeartbeat: false,
		Writer:               w,
		WriterMu:             sync.Mutex{},
		Closed:               make(chan interface{}),
	}
	defer close(client.Closed)

	for {
		msg, err := ReadMessage(r)
		if err != nil {
			if err == ErrNotImplemented {
				return client.writeMessage(MessageError{Msg: err.Error()})
			}
			return err
		}
		log.Printf("---> %T %+v", msg, msg)

		switch msg := msg.(type) {
		case MessagePlate:
			if client.Camera == nil {
				return client.writeMessage(MessageError{Msg: "must be a camera to send plates"})
			}

			ticketer.ObservePlate(*client.Camera, msg)

		case MessageWantHeartbeat:
			if client.HasSentWantHeartbeat {
				return client.writeMessage(MessageError{Msg: "cannot send WantHeartbeat twice"})
			}
			client.HasSentWantHeartbeat = true
			if msg.Interval > 0 {
				ticker := time.NewTicker(time.Duration(msg.Interval) * Decisecond)
				defer ticker.Stop()
				go client.sendHeartbeats(ticker)
			}

		case MessageIAmCamera:
			if client.Camera != nil || client.Dispatcher != nil {
				return client.writeMessage(MessageError{Msg: "cannot identify yourself twice"})
			}
			client.Camera = &msg

		case MessageIAmDispatcher:
			if client.Camera != nil || client.Dispatcher != nil {
				return client.writeMessage(MessageError{Msg: "cannot identify yourself twice"})
			}
			client.Dispatcher = &msg

			dispatcherTracker.RegisterDispatcher(msg, client)
			defer dispatcherTracker.UnregisterDispatcher(client)

		default:
			return client.writeMessage(MessageError{Msg: fmt.Sprintf("clients cannot send %T", msg)})
		}
	}
}

func (client *Client) sendHeartbeats(ticker *time.Ticker) {
	for range ticker.C {
		client.writeMessage(MessageHeartbeat{})
	}
}

func (client *Client) writeMessage(msg MessageWriter) error {
	client.WriterMu.Lock()
	defer client.WriterMu.Unlock()

	log.Printf("    <=== %T %+v", msg, msg)
	return WriteMessage(client.Writer, msg)
}

func (srv Server) Listen(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	dispatcherTracker := &DispatcherTracker{
		clients:          make(map[*Client]MessageIAmDispatcher),
		roads:            make(map[uint16]chan MessageTicket),
		dispatcherCounts: make(map[uint16]int),
		dispatcherJoined: *sync.NewCond(&sync.Mutex{}),
	}

	ticketer := &Ticketer{
		observationCh:     make(chan observation),
		dispatcherTracker: dispatcherTracker,
	}

	go ticketer.ListenPlates()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("tcp server accept error", err)
		}

		go handleConnection(ticketer, dispatcherTracker, conn)
	}
}
