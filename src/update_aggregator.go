package main

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/subpop/go-log"
)

const CountThreshold = 32
const TimeThreshold = 15 // Seconds

type UpdateAggregator struct {
	Updates   []V1Update
	Count     int
	ReturnURL string
	MessageID string
	LastSend  time.Time
}

func NewUpdateAggregator(returnUrl string, messageID string) UpdateAggregator {
	return UpdateAggregator{
		Updates:   make([]V1Update, CountThreshold),
		ReturnURL: returnUrl,
		MessageID: messageID,
		LastSend:  time.Now(),
	}
}

func (a *UpdateAggregator) Aggregate(events <-chan V1Update, c ExternalCommunication) {
	if err := c.Connect(); err != nil {
		log.Fatal(err)
	}
	defer c.Disconnect()

	for event := range events {
		a.DispatchEvent(event, c)
	}
}

func (a *UpdateAggregator) DispatchEvent(event V1Update, c ExternalCommunication) {
	a.Add(event)

	if a.Count == CountThreshold || time.Since(a.LastSend).Seconds() > TimeThreshold || event.Type == "exit" {
		a.SendUpdates(c)
	}
}

func (a *UpdateAggregator) Add(event V1Update) {
	a.Updates[a.Count] = event
	a.Count++
}

func (a *UpdateAggregator) SendUpdates(c ExternalCommunication) {
	updates := V1Updates{
		Version: "1",
		Updates: a.Updates[:a.Count],
	}

	payload, err := json.Marshal(updates)
	if err != nil {
		log.Errorf("Failed to marshal json: %v", err)
		return
	}

	data := Message{
		Directive:  a.ReturnURL,
		MessageID:  uuid.New().String(),
		ResponseTo: a.MessageID,
		Content:    payload,
		Metadata: map[string]string{
			"Content-Type": "application/json",
		},
	}

	if err := c.Send(data); err != nil {
		log.Error(err)
	}

	a.Count = 0
	a.LastSend = time.Now()
}
