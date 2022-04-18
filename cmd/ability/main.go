package main

import (
	"github.com/milobella/ability-sdk-go/pkg/ability"
)

const (
	playAction  = "play"
	pauseAction = "pause"
)

// fun main()
func main() {
	// Read configuration
	conf := ability.ReadConfiguration()

	// Initialize server
	server := ability.NewServer("ChromeCast", conf.Server.Port)
	playHandler := newChromeCastActionHandler(playAction)
	pauseHandler := newChromeCastActionHandler(pauseAction)

	// Register first the conditions on actions because they have priority on intents.
	// The condition returns true if an action is pending.
	server.RegisterRule(func(request *ability.Request) bool {
		return request.IsInSlotFillingAction(playAction)
	}, playHandler)
	server.RegisterRule(func(request *ability.Request) bool {
		return request.IsInSlotFillingAction(pauseAction)
	}, pauseHandler)

	// Then we register intents routing rules.
	// It means that if no pending action has been found in the context, we'll use intent to decide the handler.
	server.RegisterIntentRule("CHROME_CAST_PLAY", playHandler)
	server.RegisterIntentRule("CHROME_CAST_PAUSE", pauseHandler)
	server.RegisterIntentRule("PLAY", playHandler)
	server.RegisterIntentRule("PAUSE", pauseHandler)

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
	if instrument := req.InterpretInstrumentFromNLU(ability.InstrumentKindChromeCast); instrument != nil {
		// We found the instrument in NLU and return the response
		buildOneInstrumentsResponse(action, *instrument, resp)
		return
	}

	if req.IsInSlotFillingAction(action) {
		// We are in slot filling context and no answer has been found in the NLU, we stop here.
		resp.Nlg.Sentence = "I didn't find any chrome cast instrument in the device matching your request."
		return
	}

	// Build a reprompt answer
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
