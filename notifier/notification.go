package notifier

import "time"

type Notification struct {
	Timestamp time.Time `json:"timestamp"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
}
