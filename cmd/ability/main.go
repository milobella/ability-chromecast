package main

import (
	"github.com/milobella/ability-sdk-go/pkg/ability"
)

// fun main()
func main() {
	// Read configuration
	conf := ability.ReadConfiguration()

	// Initialize server
	server := ability.NewServer("ChromeCast", conf.Server.Port)

	intentsByAction := make(map[string]string, 2)
	intentsByAction["play"] = "CHROME_CAST_PLAY"
	intentsByAction["pause"] = "CHROME_CAST_PAUSE"

	handlers := make(map[string]func(_ *ability.Request, resp *ability.Response), 2)
	for action := range intentsByAction {
		handlers[action] = newChromeCastActionHandler(action)
	}

	// Register first the conditions on actions because they have priority on intents.
	// The condition returns true if an action is pending.
	for action := range intentsByAction {
		server.RegisterRule(func(request *ability.Request) bool {
			return request.IsInSlotFillingAction(action)
		}, handlers[action])
	}
	// Then we register intents routing rules.
	// It means that if no pending action has been found in the context, we'll use intent to decide the handler.
	for action, intent := range intentsByAction {
		server.RegisterIntentRule(intent, handlers[action])
	}
	server.Serve()
}

func newChromeCastActionHandler(action string) func(*ability.Request, *ability.Response) {
	return func(req *ability.Request, resp *ability.Response) {
		instruments := req.Device.CanDo(ability.InstrumentKindChromeCast, action)

		if len(instruments) <= 0 {
			// No chrome cast found, we return an error.
			resp.Nlg.Sentence = "I didn't find any chrome cast instrument in the device."
			return
		} else if len(instruments) > 1 {
			// Several chrome casts found, we apply a disambiguation algorithm
			buildSeveralInstrumentsResponse(action, instruments, req, resp)
			return
		}

		// In any other case, we found the instrument and return the response
		buildOneInstrumentsResponse(action, instruments[0], resp)
	}
}

func buildSeveralInstrumentsResponse(action string, instruments []ability.Instrument, req *ability.Request, resp *ability.Response) {
	if req.IsInSlotFillingAction(action) {
		// If the request is in a slot filling context, we try to find the instrument in the NLU.
		if instrument := req.InterpretInstrumentFromNLU(ability.InstrumentKindChromeCast); instrument != nil {
			// We found the instrument and return the response
			buildOneInstrumentsResponse(action, *instrument, resp)
		} else {
			// No chrome cast match the request, we return an error.
			resp.Nlg.Sentence = "I didn't find any chrome cast instrument in the device matching your request."
		}
		return
	}

	// Build a reprompt if we are not in slot filling context
	var instrumentsNames []string
	for _, instrument := range instruments {
		instrumentsNames = append(instrumentsNames, instrument.Name)
	}
	resp.Nlg.Sentence = "I found several chrome cast instruments in the device : {{instruments}}."
	resp.Nlg.Params = []ability.NLGParam{{
		Name:  "instruments",
		Value: instrumentsNames,
		Type:  "enumerated_list",
	}}
	resp.Context.SlotFilling.Action = action
	resp.Context.SlotFilling.MissingSlots = []string{"instrument_name"}
	resp.AutoReprompt = true
}

func buildOneInstrumentsResponse(action string, instrument ability.Instrument, resp *ability.Response) {
	resp.Nlg.Sentence = "Executing the action {{action}} on the chrome cast {{instrument}}."
	resp.Nlg.Params = []ability.NLGParam{{
		Name:  "action",
		Value: action,
		Type:  "string",
	}, {
		Name:  "instrument",
		Value: instrument.Name,
		Type:  "string",
	}}
	resp.Actions = []ability.Action{{
		Identifier: action,
		Params: []ability.ActionParameter{{
			Key:   "instrument",
			Value: instrument.Name,
		}},
	}}
}
