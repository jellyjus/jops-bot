package usecase

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"jops-bot/entity"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func makeKeyboard(answers []entity.Answer, cols int) tgbotapi.InlineKeyboardMarkup {
	if cols <= 0 {
		cols = 4
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	for i, a := range answers {
		btn := tgbotapi.NewInlineKeyboardButtonData(a.Label, a.Data)
		row = append(row, btn)
		if (i+1)%cols == 0 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func StartBot() {
	token := "8584942263:AAGmcFw6qzP5rZY1pP8X2V_yWulqlIPg6mw"

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("NewBotAPI: %v", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	store := NewStore()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("GetUpdatesChan: %v", err)
	}

	for upd := range updates {
		if upd.Message != nil {
			handleMessage(bot, store, upd.Message)
			continue
		}
		if upd.CallbackQuery != nil {
			handleCallback(bot, store, upd.CallbackQuery)
			continue
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, store *Store, m *tgbotapi.Message) {
	chatID := m.Chat.ID
	text := strings.TrimSpace(m.Text)

	switch {
	case strings.HasPrefix(text, "/start"):
		s := store.Reset(chatID)
		sendQuestion(bot, chatID, s.Step)
	case strings.HasPrefix(text, "/restart"):
		s := store.Reset(chatID)
		sendQuestion(bot, chatID, s.Step)
	default:
		// Gentle nudge
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, отвечай кнопками 🙂\n\nКоманды: /start или /restart")
		_, _ = bot.Send(msg)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, store *Store, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	s := store.GetOrCreate(chatID)

	// Always answer callback to stop Telegram client "loading"
	_, _ = bot.AnswerCallbackQuery(tgbotapi.NewCallback(cq.ID, ""))

	kind, val, ok := parseCallback(cq.Data)
	if !ok {
		return
	}

	// Validate step -> expected kind
	switch s.Step {
	case entity.StepQ1, entity.StepQ2, entity.StepQ3, entity.StepQ6, entity.StepQ7, entity.StepQ8:
		if kind != "e" {
			return
		}
		e, ok := parseElement(val)
		if !ok {
			return
		}
		s.Score[e]++

		if s.Step == entity.StepQ3 {
			s.AnswerQ3 = e
		}
		if s.Step == entity.StepQ8 {
			s.AnswerQ8 = e
		}

	case entity.StepQ4:
		// Q4 does not affect final result (by your business rules),
		// but we still validate the callback.
		if kind != "a4" {
			return
		}
		_, ok := parseArchetype(val)
		if !ok {
			return
		}

	case entity.StepQ5:
		if kind != "a5" {
			return
		}
		a, ok := parseArchetype(val)
		if !ok {
			return
		}
		s.ArchetypeQ5 = a

	default:
		// If finished, allow user to restart via /restart
		return
	}

	// Move to next
	s.Step = nextStep(s.Step)

	if s.Step == entity.StepDone {
		// Compute and send final result
		elementBlock := computeElementResult(s)
		archetypeBlock := computeArchetypeResult(s)

		final := elementBlock + "\n\n" + archetypeBlock + "\n\n" + "Хочешь пройти заново? /restart"
		msg := tgbotapi.NewMessage(chatID, final)
		msg.ParseMode = "" // plain text
		_, _ = bot.Send(msg)
		return
	}

	sendQuestion(bot, chatID, s.Step)
}

// -------------------- Optional: debug helpers --------------------

func debugScores(score map[entity.Element]int) string {
	type kv struct {
		E entity.Element
		V int
	}
	arr := []kv{
		{entity.Fire, score[entity.Fire]},
		{entity.Water, score[entity.Water]},
		{entity.Air, score[entity.Air]},
		{entity.Earth, score[entity.Earth]},
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].V > arr[j].V })
	var b strings.Builder
	for _, it := range arr {
		b.WriteString(fmt.Sprintf("%s=%d ", it.E, it.V))
	}
	return strings.TrimSpace(b.String())
}
