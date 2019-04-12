package command

import (
	"github.com/x-cellent/decs"
	"github.com/x-cellent/gokoban/event"
)

const MoveUp = "moveUp"

func NewMoveUp() decs.Command {
	return Bus.NewCommand(MoveUp, nil)
}

func newMoveUp() *decs.CommandDefinition {
	return &decs.CommandDefinition{
		Name:         MoveUp,
		UndoneEvents: []string{event.OnMoveUpUndone},
	}
}
