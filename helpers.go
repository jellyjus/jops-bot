package main

import (
	"strings"

	"jops-bot/entity"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func parseCallback(data string) (kind string, val string, ok bool) {
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func parseElement(s string) (entity.Element, bool) {
	switch s {
	case "fire":
		return entity.Fire, true
	case "water":
		return entity.Water, true
	case "air":
		return entity.Air, true
	case "earth":
		return entity.Earth, true
	default:
		return "", false
	}
}

func parseArchetype(s string) (entity.Archetype, bool) {
	switch s {
	case "resource":
		return entity.Resource, true
	case "path":
		return entity.Path, true
	case "knowledge":
		return entity.Knowledge, true
	case "power":
		return entity.Power, true
	case "will":
		return entity.Will, true
	case "feelings":
		return entity.Feelings, true
	case "home":
		return entity.Home, true
	default:
		return "", false
	}
}

func nextStep(step entity.Step) entity.Step {
	switch step {
	case entity.StepQ1:
		return entity.StepQ2
	case entity.StepQ2:
		return entity.StepQ3
	case entity.StepQ3:
		return entity.StepQ4
	case entity.StepQ4:
		return entity.StepQ5
	case entity.StepQ5:
		return entity.StepQ6
	case entity.StepQ6:
		return entity.StepQ7
	case entity.StepQ7:
		return entity.StepQ8
	case entity.StepQ8:
		return entity.StepDone
	default:
		return entity.StepDone
	}
}

var elementOrder = map[entity.Element]int{
	entity.Fire:  1,
	entity.Water: 2,
	entity.Air:   3,
	entity.Earth: 4,
}

func elementKeyForPair(a, b entity.Element) string {
	// Stable ordering for mixed text keys
	if elementOrder[a] > elementOrder[b] {
		a, b = b, a
	}
	return string(a) + "-" + string(b)
}

func computeElementResult(s *entity.Session) string {
	// Determine leaders
	max := -1
	for _, e := range []entity.Element{entity.Fire, entity.Water, entity.Air, entity.Earth} {
		if s.Score[e] > max {
			max = s.Score[e]
		}
	}
	var leaders []entity.Element
	for _, e := range []entity.Element{entity.Fire, entity.Water, entity.Air, entity.Earth} {
		if s.Score[e] == max {
			leaders = append(leaders, e)
		}
	}

	switch len(leaders) {
	case 1:
		return entity.ElementResultText[string(leaders[0])]
	case 2:
		key := elementKeyForPair(leaders[0], leaders[1])
		return entity.ElementResultText[key]
	default:
		// Triple (or more) tie: resolve by Q8, then Q3
		if s.AnswerQ8 != "" {
			return entity.ElementResultText[string(s.AnswerQ8)]
		}
		return entity.ElementResultText[string(s.AnswerQ3)]
	}
}

func computeArchetypeResult(s *entity.Session) string {
	txt, ok := entity.ArchetypeResultText[s.ArchetypeQ5]
	if !ok {
		return "Архетип не определён (нет ответа на вопрос 5)."
	}
	return txt
}

func newSession() *entity.Session {
	return &entity.Session{
		Step:  entity.StepQ1,
		Score: map[entity.Element]int{entity.Fire: 0, entity.Water: 0, entity.Air: 0, entity.Earth: 0},
	}
}

func sendQuestion(bot *tgbotapi.BotAPI, chatID int64, step entity.Step) {
	q, ok := entity.Questions[step]
	if !ok {
		msg := tgbotapi.NewMessage(chatID, "Что-то пошло не так. /restart")
		_, _ = bot.Send(msg)
		return
	}

	cols := 4
	if step == entity.StepQ4 || step == entity.StepQ5 {
		cols = 4 // will wrap automatically
	}

	msg := tgbotapi.NewMessage(chatID, q.Title)
	msg.ReplyMarkup = makeKeyboard(q.Answers, cols)
	_, _ = bot.Send(msg)
}
