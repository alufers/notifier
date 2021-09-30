package notifier

import "context"

type NotificationSink interface {
	DeliverNotification(notification *Notification) error
}

type NotificationSinkWithQuestions interface {
	NotificationSink
	AskQuestion(ctx context.Context, question *Question) (*Answer, error)
}
