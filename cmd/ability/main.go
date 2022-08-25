package main

import (
	"github.com/milobella/ability-sdk-go/pkg/config"
	"github.com/milobella/ability-sdk-go/pkg/model"
	"github.com/milobella/ability-sdk-go/pkg/server"
	"github.com/milobella/ability-sdk-go/pkg/server/interpreters"
)

const (
	playAction  = "play"
	pauseAction = "pause"
)

// fun main()
func main() {
	// Read configuration
	conf := config.Read()

	// Initialize server
	srv := server.New("ChromeCast", conf.Server.Port)
	playHandler := newChromeCastActionHandler(playAction)
	pauseHandler := newChromeCastActionHandler(pauseAction)

	// Register first the conditions on actions because they have priority on intents.
	// The condition returns true if an action is pending.
	srv.Register(server.IfInSlotFilling(playAction), playHandler)
	srv.Register(server.IfInSlotFilling(pauseAction), pauseHandler)

	// Then we register intents routing rules.
	// It means that if no pending action has been found in the context, we'll use intent to decide the handler.
	srv.Register(server.IfIntents("CHROME_CAST_PLAY", "PLAY"), playHandler)
	srv.Register(server.IfIntents("CHROME_CAST_PAUSE", "PAUSE"), pauseHandler)

	srv.Serve()
}

func newChromeCastActionHandler(action string) func(*model.Request, *model.Response) {
	return func(request *model.Request, response *model.Response) {
		instrument, stopper := interpreters.FromInstrument(model.InstrumentKindChromeCast, action).Interpret(request)
		if stopper != nil {
			stopper(response)
			return
		}

		response.Nlg.Sentence = "Executing the action {{ action }} on the chrome cast {{ instrument }}."
		response.Nlg.Params = []model.NLGParam{{
			Name:  "action",
			Value: action,
			Type:  "string",
		}, {
			Name:  "instrument",
			Value: instrument,
			Type:  "string",
		}}
		response.Actions = []model.Action{{
			Identifier: action,
			Params: []model.ActionParameter{{
				Key:   "instrument",
				Value: *instrument,
			}},
		}}
	}
}
