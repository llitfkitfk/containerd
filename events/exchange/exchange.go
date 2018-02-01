package exchange

import (
	goevents "github.com/docker/go-events"
)

// Exchange broadcasts events
type Exchange struct {
	broadcaster *goevents.Broadcaster
}

// NewExchange returns a new event Exchange
func NewExchange() *Exchange {
	return &Exchange{
		broadcaster: goevents.NewBroadcaster(),
	}
}
