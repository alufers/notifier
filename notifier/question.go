package notifier

import "time"

type QuestionKind string

var (
	QuestionKind_YesNo QuestionKind = "yesno"
	QuestionKind_Text  QuestionKind = "text"
)

type Question struct {
	Text      string       `json:"text"`
	Kind      QuestionKind `json:"kind"`
	Timestamp time.Time    `json:"timestamp"`
}

type Answer struct {
	TimedOut       bool          `json:"timedOut"`
	AnwserDuration time.Duration `json:"answerDuration" swaggertype:"primitive,integer"`
	Value          interface{}   `json:"value"`
}
