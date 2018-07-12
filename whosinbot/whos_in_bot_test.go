package whosinbot

import (
	"github.com/col/whosinbot/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockDataStore struct {
	startRollCallCalled bool
	startRollCallWith   *domain.RollCall

	endRollCallCalled bool
	endRollCallWith   *domain.RollCall

	setQuietCalled       bool
	setQuietWithRollCall *domain.RollCall
	setQuietWithBool     bool

	setResponseCalled bool
	setResponseWith   *domain.RollCallResponse

	rollCall          *domain.RollCall
	rollCallResponses []domain.RollCallResponse
}

func (d *MockDataStore) StartRollCall(rollCall domain.RollCall) error {
	d.startRollCallCalled = true
	d.startRollCallWith = &rollCall
	return nil
}

func (d *MockDataStore) EndRollCall(rollCall domain.RollCall) error {
	d.endRollCallCalled = true
	d.endRollCallWith = &rollCall
	return nil
}

func (d *MockDataStore) SetQuiet(rollCall domain.RollCall, quiet bool) error {
	d.setQuietCalled = true
	d.setQuietWithRollCall = &rollCall
	d.setQuietWithBool = quiet
	return nil
}

func (d *MockDataStore) SetResponse(rollCallResponse domain.RollCallResponse) error {
	d.setResponseCalled = true
	d.setResponseWith = &rollCallResponse
	switch rollCallResponse.Response {
	case "in":
		d.rollCall.In = append(d.rollCall.In, rollCallResponse)
	case "out":
		d.rollCall.Out = append(d.rollCall.Out, rollCallResponse)
	case "maybe":
		d.rollCall.Maybe = append(d.rollCall.Maybe, rollCallResponse)
	}
	return nil
}

func (d *MockDataStore) GetRollCall(chatID int64) (*domain.RollCall, error) {
	return d.rollCall, nil
}

func (d *MockDataStore) GetRollCallResponses(chatID int64) ([]domain.RollCallResponse, error) {
	return d.rollCallResponses, nil
}

var mockDataStore *MockDataStore
var bot *WhosInBot

func setUp() {
	mockDataStore = &MockDataStore{
		startRollCallCalled: false,
		startRollCallWith:   nil,
		endRollCallCalled:   false,
		endRollCallWith:     nil,
		setResponseCalled:   false,
		setResponseWith:     nil,
		rollCall:            nil,
		rollCallResponses:   nil,
	}
	bot = &WhosInBot{DataStore: mockDataStore}
}

func TestStartRollCall(t *testing.T) {
	setUp()
	command := domain.Command{ChatID: 123, Name: "start_roll_call", Params: []string{"sample title"}}
	response, err := bot.HandleCommand(command)
	// Validate data store
	assert.True(t, mockDataStore.startRollCallCalled)
	assert.NotNil(t, mockDataStore.startRollCallWith)
	assert.Equal(t, int64(123), mockDataStore.startRollCallWith.ChatID)
	assert.Equal(t, "sample title", mockDataStore.startRollCallWith.Title)
	// Validate response
	assertBotResponse(t, response, err, 123, "Roll call started", nil)
}

func TestEndRollCallWhenRollCallExists(t *testing.T) {
	setUp()
	mockDataStore.rollCall = &domain.RollCall{ChatID: 123, Title: ""}
	command := domain.Command{ChatID: 123, Name: "end_roll_call", Params: []string{}}
	response, err := bot.HandleCommand(command)
	// Validate data store
	assert.True(t, mockDataStore.endRollCallCalled)
	assert.NotNil(t, mockDataStore.endRollCallWith)
	assert.Equal(t, int64(123), mockDataStore.endRollCallWith.ChatID)
	// Validate response
	assertBotResponse(t, response, err, 123, "Roll call ended", nil)
}

func TestEndRollCallWhenRollCallDoesNotExists(t *testing.T) {
	setUp()
	mockDataStore.rollCall = nil
	command := domain.Command{ChatID: 123, Name: "end_roll_call", Params: []string{}}
	response, err := bot.HandleCommand(command)
	// Validate data store
	assert.False(t, mockDataStore.endRollCallCalled)
	// Validate response
	assertBotResponse(t, response, err, 123, "No roll call in progress", nil)
}

func TestInWhenNoRollCallInProgress(t *testing.T) {
	setUp()
	response, err := bot.HandleCommand(responseCommand("in"))
	assertBotResponse(t, response, err, 123, "No roll call in progress", nil)
}

func TestInWhenRollCallIsInProgress(t *testing.T) {
	setUp()
	mockDataStore.rollCall = &domain.RollCall{ChatID: 123, Title: ""}

	response, err := bot.HandleCommand(responseCommand("in"))

	assertResponsePersisted(t, 123, "in", "JohnSmith")
	assertBotResponse(t, response, err, 123, "1. JohnSmith (sample reason)", nil)
}

func TestOutWhenNoRollCallInProgress(t *testing.T) {
	setUp()
	response, err := bot.HandleCommand(responseCommand("out"))
	assert.False(t, mockDataStore.setResponseCalled)
	assertBotResponse(t, response, err, 123, "No roll call in progress", nil)
}

func TestOutWhenRollCallIsInProgress(t *testing.T) {
	setUp()
	mockDataStore.rollCall = &domain.RollCall{ChatID: 123, Title: ""}
	response, err := bot.HandleCommand(responseCommand("out"))
	assertResponsePersisted(t, 123, "out", "JohnSmith")
	assertBotResponse(t, response, err, 123, "Out\n - JohnSmith (sample reason)", nil)
}

func TestMaybeWhenNoRollCallInProgress(t *testing.T) {
	setUp()
	response, err := bot.HandleCommand(responseCommand("maybe"))
	assert.False(t, mockDataStore.setResponseCalled)
	assertBotResponse(t, response, err, 123, "No roll call in progress", nil)
}

func TestMaybeWhenRollCallIsInProgress(t *testing.T) {
	setUp()
	mockDataStore.rollCall = &domain.RollCall{ChatID: 123, Title: ""}
	response, err := bot.HandleCommand(responseCommand("maybe"))
	assertResponsePersisted(t, 123, "maybe", "JohnSmith")
	assertBotResponse(t, response, err, 123, "Maybe\n - JohnSmith (sample reason)", nil)
}

func TestWhosIn(t *testing.T) {
	setUp()
	mockDataStore.rollCall = &domain.RollCall{
		ChatID: 123,
		Title:  "Test Title",
		In: []domain.RollCallResponse{
			{ChatID: 123, UserID: 1, Name: "User 1", Response: "in", Reason: ""},
		},
		Out: []domain.RollCallResponse{
			{ChatID: 123, UserID: 1, Name: "User 2", Response: "out", Reason: ""},
		},
		Maybe: []domain.RollCallResponse{
			{ChatID: 123, UserID: 1, Name: "User 3", Response: "maybe", Reason: ""},
		},
	}
	response, err := bot.HandleCommand(responseCommand("whos_in"))
	assertBotResponse(t, response, err, 123, "Test Title\n1. User 1\n\nOut\n - User 2\n\nMaybe\n - User 3", nil)
}

func TestShh(t *testing.T) {
	setUp()
	mockDataStore.rollCall = &domain.RollCall{
		ChatID: 123,
		Title:  "Test Title",
		Quiet:  false,
	}

	response, err := bot.HandleCommand(command("shh", nil))
	// Validate data store
	assert.True(t, mockDataStore.setQuietCalled)
	assert.NotNil(t, mockDataStore.setQuietWithRollCall)
	assert.True(t, mockDataStore.setQuietWithBool)
	// Validate Response
	assertBotResponse(t, response, err, 123, "Ok fine, I'll be quiet. 🤐", nil)
}

func TestShhWithNRollCallInProgress(t *testing.T) {
	setUp()
	response, err := bot.HandleCommand(command("shh", nil))
	assertBotResponse(t, response, err, 123, "No roll call in progress", nil)
}

// Test Helpers
func command(name string, params []string) domain.Command {
	return domain.Command{
		ChatID: 123,
		Name:   name,
		Params: params,
		From:   domain.User{UserID: 456, Username: "JohnSmith"},
	}
}

func responseCommand(status string) domain.Command {
	return domain.Command{
		ChatID: 123,
		Name:   status,
		Params: []string{"sample reason"},
		From:   domain.User{UserID: 456, Username: "JohnSmith"},
	}
}

func assertResponsePersisted(t *testing.T, chatID int, status string, name string) {
	assert.True(t, mockDataStore.setResponseCalled, "should call setResponse")
	assert.NotNil(t, mockDataStore.setResponseWith)
	if mockDataStore.setResponseWith != nil {
		assert.Equal(t, int64(chatID), mockDataStore.setResponseWith.ChatID)
		assert.Equal(t, status, mockDataStore.setResponseWith.Response)
		assert.Equal(t, name, mockDataStore.setResponseWith.Name)
	}
}

func assertBotResponse(t *testing.T, response *domain.Response, err error, chatID int, text string, error error) {
	assert.Equal(t, int64(chatID), response.ChatID)
	assert.Equal(t, text, response.Text)
	assert.Equal(t, error, err)
}
