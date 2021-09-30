package notifier

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramNotificationSink struct {
	Name            string
	BotToken        string
	ChatID          int64
	TelegramManager *TelegramManager
	bot             *tgbotapi.BotAPI
}

func (sink *TelegramNotificationSink) Init() error {
	bot, err := sink.TelegramManager.RegisterBot(sink.BotToken)
	if err != nil {
		return err
	}

	sink.bot = bot
	log.Printf("Successfully initialized %T", *sink)

	return nil
}

func (sink *TelegramNotificationSink) DeliverNotification(notification *Notification) error {
	titleStr := ""
	if notification.Title != "" {
		titleStr = fmt.Sprintf("<b>%v</b>\n", notification.Title)
	}
	msg := tgbotapi.NewMessage(sink.ChatID, fmt.Sprintf(
		"%v%s\n<code>%v</code>",
		titleStr,
		notification.Body,
		formatDate(notification.Timestamp),
	))
	msg.ParseMode = "HTML"
	_, err := sink.bot.Send(msg)
	return err
}

func (sink *TelegramNotificationSink) AskQuestion(ctx context.Context, question *Question) (*Answer, error) {
	switch question.Kind {
	case QuestionKind_YesNo:
		return sink.askYesNoQuestion(ctx, question)
	default:
		return nil, fmt.Errorf("unsupported question kind: %v", question.Kind)
	}

}

func (sink *TelegramNotificationSink) askYesNoQuestion(ctx context.Context, question *Question) (*Answer, error) {
	questionID := fmt.Sprintf("%x", rand.Int63())
	msg := tgbotapi.NewMessage(sink.ChatID, fmt.Sprintf("%v\n<code>%v</code>", question.Text, formatDate(question.Timestamp)))
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", "yes_"+questionID),
			tgbotapi.NewInlineKeyboardButtonData("No", "no_"+questionID),
		),
	)

	msgSent, err := sink.bot.Send(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send question: %v", err)
	}
	questionAskedTime := time.Now()
	answerChan := make(chan *Answer)
	removeListener := sink.TelegramManager.AddUpdateListener(sink.BotToken, func(update *tgbotapi.Update) {
		if update.CallbackQuery == nil {
			return
		}
		if update.CallbackQuery.Message.Chat.ID != sink.ChatID {
			return
		}
		switch update.CallbackQuery.Data {
		case "yes_" + questionID:
			_, err := sink.bot.Send(tgbotapi.NewEditMessageReplyMarkup(
				msgSent.Chat.ID,
				msgSent.MessageID,
				tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Answered: Yes", "i"),
					),
				),
			))
			if err != nil {
				log.Printf("failed to edit message: %v", err)
			}
			sink.bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Yes"))
			answerChan <- &Answer{
				TimedOut:       false,
				Value:          true,
				AnwserDuration: time.Since(questionAskedTime),
			}
		case "no_" + questionID:
			sink.bot.Send(tgbotapi.NewEditMessageReplyMarkup(
				msgSent.Chat.ID,
				msgSent.MessageID,
				tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Answered: No", "i"),
					),
				),
			))
			sink.bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "No"))
			answerChan <- &Answer{
				TimedOut:       false,
				Value:          false,
				AnwserDuration: time.Since(questionAskedTime),
			}
		}
	})
	select {
	case answer := <-answerChan:
		removeListener()
		return answer, nil
	case <-ctx.Done():
		removeListener()
		_, err := sink.bot.Send(tgbotapi.NewEditMessageReplyMarkup(
			msgSent.Chat.ID,
			msgSent.MessageID,
			tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Timed out (No)", "i"),
				),
			),
		))
		if err != nil {
			log.Printf("failed to edit message after question timeout: %v", err)
		}
		return &Answer{
			Value:          false,
			TimedOut:       true,
			AnwserDuration: time.Since(questionAskedTime),
		}, nil

	}

}
