package botstate

import (
	"encoding/json"
	"errors"
)

//State are used to save states
//Name is the identifier of each state, so it must be unique.
//Executes is the method to be executed when bot call state.
//Callback is the method to be executed in the next state, commonly used to receive input data and validate it.
//Next is the identifier to next state, it is added to user data after the call of Executes method.
type State struct {
	Name     string
	Executes func(bot *Bot) bool
	Callback func(bot *Bot) bool
	Next     string
}

//Bot are used to initialize states and control bot flow.
//The methods in Executes and Callback in the states receive an instance of Bot struct
type Bot struct {
	States []State
	Data   *BotData
}

//New returns new Bot struct with BotData
func New(states []State) *Bot {
	return &Bot{
		States: states,
		Data:   &BotData{},
	}
}

//ExecuteState define the current_state using argument name in the user's data.
//
//If exists callback in user's data (state_with_callback), execute it first with the executeCallback method.
//Terminates execution if callback returns false.
//
//If there is no callback, the flow continues and if the current state has a method in the item State.Callback, this value will be defined in the user's current data (state_with_callback) to be executed later.
//
//After all checks, the method in State.Executes is executed.
//The current state is defined using the value of State.Next if execution return true.
//
//Return execution boolean and error if exists.
func (b *Bot) ExecuteState(name string) (bool, error) {
	for _, state := range b.States {
		if state.Name == name {
			if b.Data.UserID == "" {
				return false, errors.New("Undefined user to execute state " + state.Name + ".")
			}

			if state.Executes == nil {
				return false, errors.New("Method to execute in the " + state.Name + " state is nil.")
			}

			b.Data.SetCurrentState(state.Name)

			callbackResp, err := b.executeCallback()

			if err != nil {
				return false, err
			}

			if callbackResp == false {
				return false, nil
			}

			if state.Callback != nil {
				err := b.Data.SetStateWithCallback(state.Name)

				if err != nil {
					return false, err
				}
			}

			execute := state.Executes(b)

			if execute == true && state.Next != "" {
				b.Data.SetCurrentState(state.Next)
			}

			return execute, nil
		}
	}

	return false, errors.New("No state to execute with name " + name + ".")
}

//executeCallback will get state_with_callback from user's data.
//And execute the executeCallbackFromState method passing state name as argument.
//Return callback execution boolean and error if exists.
func (b *Bot) executeCallback() (bool, error) {
	stateWithCallback, _ := b.Data.GetStateWithCallback()

	if stateWithCallback != "" {
		return b.executeCallbackFromState(stateWithCallback), nil
	}

	return true, nil
}

//executeCallbackFromState will loop through all states to find the state with argument name.
//Checks if the state has method in State.Callback.
//Execute method in State.Callback.
//Return callback boolean response.
func (b *Bot) executeCallbackFromState(name string) bool {
	for _, state := range b.States {
		if state.Name == name {
			if state.Callback != nil {
				return state.Callback(b)
			}
		}
	}

	return true
}

//AddMessage save messages to user data
//Can be used to return messages after bot execution
func (b *Bot) AddMessage(messages []string) error {
	if len(messages) <= 0 {
		return errors.New("undefined messages")
	}

	messages = append(b.GetMessages(), messages...)

	j, err := json.Marshal(messages)

	if err != nil {
		return err
	}

	err = b.Data.SetData(Data{
		"messages": string(j),
	})

	if err != nil {
		return err
	}

	return nil
}

//GetMessages return all messages saved in user data
//This messages is from bot, saved during execution flow
//Calling GetMessages will return messages and remove from data
func (b *Bot) GetMessages() []string {
	var messages []string

	m, _ := b.Data.Current["messages"]

	err := json.Unmarshal([]byte(m), &messages)

	b.Data.SetData(Data{
		"messages": "",
	})

	if err == nil {
		return messages
	}

	return []string{}
}
