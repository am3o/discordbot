package port

import "context"

type Service interface {
	Execute(string) (string, error)
}

type EventHandler struct {
	Service Service
}

func NewEventHandler(service Service) EventHandler {
	return EventHandler{
		Service: service,
	}
}

func (e *EventHandler) Publish(ctx context.Context, authorID string, message string) {
	_, _ = e.Service.Execute(message)
}
