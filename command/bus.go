package command

import (
	"github.com/x-cellent/decs"
	"go.uber.org/zap"
)

var Bus = decs.NewDefaultCommandBus(10 * decs.MegaByte)

func InitiateBus() {
	Bus.SetLogger(zap.NewNop())
	Bus.DefineCommands(createCommandDefinitions()...)
	Bus.HandleLocalCommandsOnDemand()
}

func createCommandDefinitions() []*decs.CommandDefinition {
	return []*decs.CommandDefinition{
		newMoveUp(),
		newMoveRight(),
		newMoveDown(),
		newMoveLeft(),
	}
}
