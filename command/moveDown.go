package command

import (
	"github.com/x-cellent/decs"
	"github.com/x-cellent/gokoban/event"
)

const MoveDown = "moveDown"

func NewMoveDown() decs.Command {
	return Bus.NewCommand(MoveDown, nil)
}

func newMoveDown() *decs.CommandDefinition {
	return &decs.CommandDefinition{
		Name:         MoveDown,
		UndoneEvents: []string{event.OnMoveDownUndone},
	}
}
