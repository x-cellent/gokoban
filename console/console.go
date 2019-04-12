package console

import (
	"flag"
	"github.com/jroimartin/gocui"
	"github.com/x-cellent/decs"
	"github.com/x-cellent/gokoban/command"
	"github.com/x-cellent/gokoban/event"
	"github.com/x-cellent/gokoban/gokoban"
	"log"
	"time"
)

const (
	retryInterval = 30 * time.Second
	flushDelay    = 20 * time.Millisecond
)

var (
	nsqdTcpAddress  = flag.String("t", "", "NSQD TCP address")
	nsqdHttpAddress = flag.String("h", "", "NSQD HTTP address")
	natsURL         = flag.String("n", "", "NATS URL")
	natsClusterID   = flag.String("c", "", "NATS cluster ID")
	natsClientID    = flag.String("C", "", "NATS client ID")
)

func init() {
	flag.StringVar(nsqdTcpAddress, "nsqd-tcp-address", "", "NSQD TCP address")
	flag.StringVar(nsqdHttpAddress, "nsqd-http-address", "", "NSQD HTTP address")
	flag.StringVar(natsURL, "nats-url", "", "NATS URL")
	flag.StringVar(natsClusterID, "nats-cluster-id", "", "NATS cluster ID")
	flag.StringVar(natsClientID, "nats-client-id", "", "NATS client ID")

	flag.Parse()

	if len(*nsqdTcpAddress) > 0 {
		command.Bus.ConfigureNsqProvider(retryInterval, flushDelay, *nsqdTcpAddress, *nsqdHttpAddress, flag.Args()...)
		return
	}

	if len(*natsURL) > 0 {
		if len(*natsClusterID) > 0 {
			command.Bus.ConfigureNatsStreamingProvider(retryInterval, flushDelay, *natsURL, *natsClusterID, *natsClientID)
			return
		}
		command.Bus.ConfigureNatsProvider(retryInterval, flushDelay, *natsURL)
		return
	}

	command.Bus.ConfigureLocalProvider(retryInterval, flushDelay)
}

func Run() {
	gui := gocui.NewGui()
	defer func() {
		gui.Cursor = true
		gui.Close()
	}()

	err := gui.Init()
	if err != nil {
		log.Panicln(err)
	}

	gui.Cursor = true

	game := newGame("gokoban/levels", 1, gui)

	gui.SetLayout(game.layout)

	gui.Cursor = false

	initiateCommandBus(game)

	if err := game.keyBindings(); err != nil {
		log.Panicln(err)
	}

	game.loadLevel()

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func initiateCommandBus(g *game) {
	command.InitiateBus()

	command.Bus.RegisterCommandHandler(command.MoveUp, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		g.movePlayer(gokoban.Up)
		notifier.NotifySuccess(event.OnMovedUp, nil)
	})
	command.Bus.RegisterCommandHandler(command.MoveRight, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		g.movePlayer(gokoban.Right)
		notifier.NotifySuccess(event.OnMovedRight, nil)
	})
	command.Bus.RegisterCommandHandler(command.MoveDown, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		g.movePlayer(gokoban.Down)
		notifier.NotifySuccess(event.OnMovedDown, nil)
	})
	command.Bus.RegisterCommandHandler(command.MoveLeft, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		g.movePlayer(gokoban.Left)
		notifier.NotifySuccess(event.OnMovedLeft, nil)
	})

	command.Bus.SubscribeAfter(decs.PurgeApplication, func(data interface{}, dispatcher decs.EventDispatcher) {
		g.reset()
	})

	command.Bus.SubscribeAfterSuccess(event.OnMovedUp, func(data interface{}, dispatcher decs.EventDispatcher) {
		g.update()
	})
	command.Bus.SubscribeAfterSuccess(event.OnMovedRight, func(data interface{}, dispatcher decs.EventDispatcher) {
		g.update()
	})
	command.Bus.SubscribeAfterSuccess(event.OnMovedDown, func(data interface{}, dispatcher decs.EventDispatcher) {
		g.update()
	})
	command.Bus.SubscribeAfterSuccess(event.OnMovedLeft, func(data interface{}, dispatcher decs.EventDispatcher) {
		g.update()
	})

	command.Bus.RegisterUndoHandler(command.MoveUp, func(cmd decs.Command, delegate decs.Delegate) {
		g.undoLastMove()
	})
	command.Bus.RegisterUndoHandler(command.MoveRight, func(cmd decs.Command, delegate decs.Delegate) {
		g.undoLastMove()
	})
	command.Bus.RegisterUndoHandler(command.MoveDown, func(cmd decs.Command, delegate decs.Delegate) {
		g.undoLastMove()
	})
	command.Bus.RegisterUndoHandler(command.MoveLeft, func(cmd decs.Command, delegate decs.Delegate) {
		g.undoLastMove()
	})
}
