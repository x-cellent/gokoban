package command

import (
	"github.com/x-cellent/decs"
	"github.com/x-cellent/gokoban/event"
)

const MoveLeft = "moveLeft"

func NewMoveLeft() decs.Command {
	return Bus.NewCommand(MoveLeft, nil)
}

func newMoveLeft() *decs.CommandDefinition {
	return &decs.CommandDefinition{
		Name:         MoveLeft,
		UndoneEvents: []string{event.OnMoveLeftUndone},
	}
}
