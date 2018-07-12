package whosinbot

import (
	"log"
	"errors"
	"github.com/col/whosinbot/domain"
	"strings"
	"fmt"
)

type WhosInBot struct {
	DataStore domain.DataStore
}

func (b *WhosInBot) HandleCommand(command domain.Command) (*domain.Response, error) {
	log.Printf("Command: %v", command.Name)
	switch command.Name {
	case "start_roll_call":
		return b.handleStart(command)
	case "end_roll_call":
		return b.handleEnd(command)
	case "in":
		return b.handleResponse(command, "in")
	case "out":
		return b.handleResponse(command, "out")
	case "maybe":
		return b.handleResponse(command, "maybe")
	case "whos_in":
		return b.handleWhosIn(command)
	case "shh":
		return b.handleShh(command)
	default:
		return nil, errors.New("Not a bot command")
	}
}

func (b *WhosInBot) handleStart(command domain.Command) (*domain.Response, error) {
	roll_call := domain.RollCall{
		ChatID: command.ChatID,
		Title:  strings.Join(command.Params, " "),
	}
	err := b.DataStore.StartRollCall(roll_call)
	if err != nil {
		return nil, err
	}
	return &domain.Response{ChatID: command.ChatID, Text: "Roll call started"}, nil
}

func (b *WhosInBot) handleEnd(command domain.Command) (*domain.Response, error) {
	rollCall, err := b.DataStore.GetRollCall(command.ChatID)
	if err != nil {
		return nil, err
	}
	if rollCall == nil {
		return &domain.Response{Text: "No roll call in progress", ChatID: command.ChatID}, nil
	}
	err = b.DataStore.EndRollCall(*rollCall)
	if err != nil {
		return nil, err
	}
	return &domain.Response{ChatID: command.ChatID, Text: "Roll call ended"}, nil
}

func (b *WhosInBot) handleShh(command domain.Command) (*domain.Response, error) {
	rollCall, err := b.DataStore.GetRollCall(command.ChatID)
	if err != nil {
		return nil, err
	}
	if rollCall == nil {
		return &domain.Response{Text: "No roll call in progress", ChatID: command.ChatID}, nil
	}
	err = b.DataStore.SetQuiet(*rollCall, true)
	if err != nil {
		return nil, err
	}
	return &domain.Response{ChatID: command.ChatID, Text: "Ok fine, I'll be quiet. 🤐"}, nil
}

func (b *WhosInBot) handleWhosIn(command domain.Command) (*domain.Response, error) {
	rollCall, err := b.DataStore.GetRollCall(command.ChatID)
	if err != nil {
		return nil, err
	}
	if rollCall == nil {
		return &domain.Response{Text: "No roll call in progress", ChatID: command.ChatID}, nil
	}
	return &domain.Response{ChatID: command.ChatID, Text: responsesList(rollCall)}, nil
}

func (b *WhosInBot) handleResponse(command domain.Command, status string) (*domain.Response, error) {
	rollCall, err := b.DataStore.GetRollCall(command.ChatID)
	if err != nil {
		return nil, err
	}
	if rollCall == nil {
		return &domain.Response{Text: "No roll call in progress", ChatID: command.ChatID}, nil
	}

	rollCallResponse := domain.RollCallResponse{
		ChatID:   command.ChatID,
		UserID:   command.From.UserID,
		Name:     command.From.Username,
		Response: status,
		Reason:   command.ParamsString(),
	}
	b.DataStore.SetResponse(rollCallResponse)

	return &domain.Response{ChatID: command.ChatID, Text: responsesList(rollCall)}, nil
}

func (b *WhosInBot) handleOut(command domain.Command) (*domain.Response, error) {
	return &domain.Response{}, nil
}

func responsesList(rollCall *domain.RollCall) (string) {
	var text = ""

	if len(rollCall.Title) > 0 {
		text += rollCall.Title
	}

	if len(rollCall.In) > 0 && len(text) > 0 {
		text += "\n"
	}
	for index, response := range rollCall.In {
		text += fmt.Sprintf("%d. %v", index+1, response.Name)
		if len(response.Reason) > 0 {
			text += fmt.Sprintf(" (%v)", response.Reason)
		}
		if index + 1 < len(rollCall.In) {
			text += "\n"
		}
	}

	text = appendResponses(text, rollCall.Out, "Out")
	text = appendResponses(text, rollCall.Maybe, "Maybe")

	return text
}

func appendResponses(text string, responses []domain.RollCallResponse, status string) (string) {
	if len(responses) > 0 {
		if len(text) > 0 {
			text += "\n\n"
		}
		text += fmt.Sprintf("%v\n", status)
	}
	for index, response := range responses {
		text += fmt.Sprintf(" - %v", response.Name)
		if len(response.Reason) > 0 {
			text += fmt.Sprintf(" (%v)", response.Reason)
		}
		if index + 1 < len(responses) {
			text += "\n"
		}
	}
	return text
}