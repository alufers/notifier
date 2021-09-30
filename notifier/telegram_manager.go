package notifier

import (
	"log"
	"math/rand"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type updateListener struct {
	handler func(update *tgbotapi.Update)
	ID      int64
}

type TelegramManager struct {
	botToTokens     map[string]*tgbotapi.BotAPI
	botsMutex       sync.RWMutex
	listenersMutex  sync.RWMutex
	updateListeners map[string][]*updateListener
}

func NewTelegramManager() *TelegramManager {
	return &TelegramManager{
		botToTokens:     make(map[string]*tgbotapi.BotAPI),
		updateListeners: make(map[string][]*updateListener),
	}
}

func (t *TelegramManager) RegisterBot(botToken string) (*tgbotapi.BotAPI, error) {
	t.botsMutex.Lock()
	defer t.botsMutex.Unlock()

	if bot, ok := t.botToTokens[botToken]; ok {
		return bot, nil
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	t.botToTokens[botToken] = bot

	go t.runUpdatesLoop(bot)

	return bot, nil
}

func (t *TelegramManager) GetBot(botToken string) *tgbotapi.BotAPI {
	t.botsMutex.RLock()
	defer t.botsMutex.RUnlock()

	return t.botToTokens[botToken]
}

func (t *TelegramManager) AddUpdateListener(botToken string, listener func(update *tgbotapi.Update)) func() {
	t.listenersMutex.Lock()
	defer t.listenersMutex.Unlock()
	added := &updateListener{
		handler: listener,
		ID:      rand.Int63(),
	}
	t.updateListeners[botToken] = append(t.updateListeners[botToken], added)
	return func() {
		t.listenersMutex.Lock()
		defer t.listenersMutex.Unlock()
		for i, l := range t.updateListeners[botToken] {
			if l.ID == added.ID {
				t.updateListeners[botToken] = append(t.updateListeners[botToken][:i], t.updateListeners[botToken][i+1:]...)
				return
			}
		}
	}
}

func (t *TelegramManager) runUpdatesLoop(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("failed to GetUpdates for bot: %v", err)
		return
	}

	for update := range updates {
		func() {
			t.listenersMutex.RLock()
			defer t.listenersMutex.RUnlock()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in telegram bot update handler: %v", r)
				}
			}()
			for _, l := range t.updateListeners[bot.Token] {
				l.handler(&update)
			}
		}()
	}
}
