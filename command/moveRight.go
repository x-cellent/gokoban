package command

import (
	"github.com/x-cellent/decs"
	"github.com/x-cellent/gokoban/event"
)

const MoveRight = "moveRight"

func NewMoveRight() decs.Command {
	return Bus.NewCommand(MoveRight, nil)
}

func newMoveRight() *decs.CommandDefinition {
	return &decs.CommandDefinition{
		Name:         MoveRight,
		UndoneEvents: []string{event.OnMoveRightUndone},
	}
}
