package speeddaemon

import (
	"math"
)

type Ticketer struct {
	observationCh     chan observation
	dispatcherTracker *DispatcherTracker
}

type observation struct {
	MessageIAmCamera
	MessagePlate
}

func (tk *Ticketer) ObservePlate(camera MessageIAmCamera, plate MessagePlate) {
	tk.observationCh <- observation{
		MessageIAmCamera: camera,
		MessagePlate:     plate,
	}
}

type carObservations map[string][]observation
type ticketCalendar map[uint32]bool

func (tkr *Ticketer) ListenPlates() {
	roads := make(map[uint16]carObservations)
	calendars := make(map[string]ticketCalendar)

	for {
		obs := <-tkr.observationCh

		road, ok := roads[obs.Road]
		if !ok {
			road = make(carObservations)
			roads[obs.Road] = road
		}

		calendar, ok := calendars[obs.Plate]
		if !ok {
			calendar = make(ticketCalendar)
			calendars[obs.Plate] = calendar
		}

		for _, ticket := range computePotentialTickets(road[obs.Plate], obs) {
			day1 := ticket.Timestamp1 / 86400
			day2 := ticket.Timestamp2 / 86400

			if alreadyIssuedTicket(calendar, day1, day2) {
				continue
			}

			go tkr.dispatcherTracker.IssueTicket(ticket)

			for i := day1; i <= day2; i++ {
				calendar[i] = true
			}
		}

		road[obs.Plate] = append(road[obs.Plate], obs)
	}
}

func computePotentialTickets(prior []observation, current observation) []MessageTicket {
	tickets := make([]MessageTicket, 0)

	for _, p := range prior {
		mile1 := p.Mile
		timestamp1 := p.Timestamp
		mile2 := current.Mile
		timestamp2 := current.Timestamp
		if timestamp2 < timestamp1 {
			mile1, mile2 = mile2, mile1
			timestamp1, timestamp2 = timestamp2, timestamp1
		}

		speed := math.Abs((float64(mile2) - float64(mile1)) / (float64(timestamp2) - float64(timestamp1)) * 60 * 60)

		if speed-float64(current.Limit) >= 0.5 {
			tickets = append(tickets, MessageTicket{
				Plate:      current.Plate,
				Road:       current.Road,
				Mile1:      mile1,
				Timestamp1: timestamp1,
				Mile2:      mile2,
				Timestamp2: timestamp2,
				SpeedX100:  uint16(speed * 100),
			})
		}
	}

	return tickets
}

func alreadyIssuedTicket(calendar ticketCalendar, day1, day2 uint32) bool {
	for i := day1; i <= day2; i++ {
		if calendar[i] {
			return true
		}
	}
	return false
}
