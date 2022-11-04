package speeddaemon

import (
	"log"
	"sync"
)

type DispatcherTracker struct {
	clients          map[*Client]MessageIAmDispatcher
	roads            map[uint16]chan MessageTicket
	dispatcherCounts map[uint16]int
	mu               sync.Mutex
	dispatcherJoined sync.Cond
}

func (dt *DispatcherTracker) RegisterDispatcher(dispatcher MessageIAmDispatcher, client *Client) {
	dt.mu.Lock()
	dt.clients[client] = dispatcher
	for _, road := range dispatcher.Roads {
		dt.dispatcherCounts[road]++
	}
	dt.dispatcherJoined.Broadcast()
	dt.mu.Unlock()
}

func (dt *DispatcherTracker) UnregisterDispatcher(client *Client) {
	dt.mu.Lock()
	dispatcher := dt.clients[client]
	delete(dt.clients, client)
	for _, road := range dispatcher.Roads {
		dt.dispatcherCounts[road]--
	}
	dt.mu.Unlock()
}

func (dt *DispatcherTracker) IssueTicket(msg MessageTicket) {
	dt.mu.Lock()
	roadCh, ok := dt.roads[msg.Road]
	if !ok {
		roadCh = make(chan MessageTicket)
		dt.roads[msg.Road] = roadCh
		go dt.listenRoad(msg.Road, roadCh)
	}
	dt.mu.Unlock()

	roadCh <- msg
}

// Yeah this sucks but at least the DispatcherTracker API is nice?
func (dt *DispatcherTracker) listenRoad(road uint16, ch chan MessageTicket) {
	var client *Client

	for {
		if client == nil {
			dt.dispatcherJoined.L.Lock()
			for dt.dispatcherCounts[road] == 0 {
				dt.dispatcherJoined.Wait()
			}
			dt.dispatcherJoined.L.Unlock()

			dt.mu.Lock()
			if dt.dispatcherCounts[road] > 0 {
				for cl, msg := range dt.clients {
					for _, droad := range msg.Roads {
						if droad == road {
							client = cl
						}
					}
				}
			}
			dt.mu.Unlock()
		} else {
			select {
			case <-client.Closed:
				client = nil
			case msg := <-ch:
				err := client.writeMessage(msg)
				if err != nil {
					log.Println(err)
					client = nil
				}
			}
		}
	}
}
