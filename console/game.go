package console

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/x-cellent/gokoban/command"
	"github.com/x-cellent/gokoban/gokoban"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	defaultReplaySpeed   = 100
	replaySpeedDecrement = 20
)

const (
	black   = "\u001b[30m"
	red     = "\u001b[31m"
	green   = "\u001b[32m"
	yellow  = "\u001b[33m"
	blue    = "\u001b[34m"
	magenta = "\u001b[35m"
	cyan    = "\u001b[36m"
	white   = "\u001b[37m"

	bgBlack   = "\u001b[40m"
	bgRed     = "\u001b[41m"
	bgGreen   = "\u001b[42m"
	bgYellow  = "\u001b[43m"
	bgBlue    = "\u001b[44m"
	bgMagenta = "\u001b[45m"
	bgCyan    = "\u001b[46m"
	bgWhite   = "\u001b[47m"

	brightBlack   = "\u001b[90m"
	brightRed     = "\u001b[91m"
	brightGreen   = "\u001b[92m"
	brightYellow  = "\u001b[93m"
	brightBlue    = "\u001b[94m"
	brightMagenta = "\u001b[95m"
	brightCyan    = "\u001b[96m"
	brightWhite   = "\u001b[97m"

	bgBrightBlack   = "\u001b[100m"
	bgBrightRed     = "\u001b[101m"
	bgBrightGreen   = "\u001b[102m"
	bgBrightYellow  = "\u001b[103m"
	bgBrightBlue    = "\u001b[104m"
	bgBrightMagenta = "\u001b[105m"
	bgBrightCyan    = "\u001b[106m"
	bgBrightWhite   = "\u001b[107m"

	reset = "\u001b[0m"
)

const (
	targetColor = green
	boxColor    = bgBlue
	brickColor  = bgYellow
	playerColor = bgWhite
)

type game struct {
	dir          string
	lvl          int
	level        *gokoban.Level
	gui          *gocui.Gui
	view         string
	replaySpeed  uint
	replayPaused bool
	replayIndex  int
}

func newGame(dir string, level int, gui *gocui.Gui) *game {
	return &game{
		dir:  dir,
		lvl:  level,
		gui:  gui,
		view: "game",
	}
}

func (g *game) maxLevel() int {
	files, err := ioutil.ReadDir(g.dir)
	if err != nil {
		return 0
	}

	max := 0
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "level") || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}
		if lvl, err := strconv.Atoi(file.Name()[5 : len(file.Name())-4]); err == nil && max < lvl {
			max = lvl
		}
	}

	return max
}

func (g *game) loadLevel() {
	g.level = gokoban.NewLevel(
		fmt.Sprintf("%s/level%d.txt", g.dir, g.lvl),
		fmt.Sprintf("%s/solution%d.txt", g.dir, g.lvl),
	)
	err := g.layout(g.gui)
	if err != nil {
		panic(err)
	}
	g.reset()
	g.update()
}

func (g *game) replaying() bool {
	return g.replaySpeed > 0
}

func (g *game) canSpeedUpReplay() bool {
	return g.replaySpeed > replaySpeedDecrement
}

func (g *game) canSlowDownReplay() bool {
	return g.replaySpeed < 1000
}

func (g *game) arrowUpHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() {
		if g.canSpeedUpReplay() {
			g.replaySpeed -= replaySpeedDecrement
		}
		return nil
	}
	if g.level.Completed() {
		return nil
	}
	if v != nil && g.level.CanMoveUp() {
		command.Bus.Do(command.NewMoveUp())
	}
	return nil
}

func (g *game) arrowRightHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() {
		if g.canSpeedUpReplay() {
			g.replaySpeed -= replaySpeedDecrement
		}
		return nil
	}
	if g.level.Completed() {
		return nil
	}
	if v != nil && g.level.CanMoveRight() {
		command.Bus.Do(command.NewMoveRight())
	}
	return nil
}

func (g *game) arrowDownHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() {
		if g.canSlowDownReplay() {
			g.replaySpeed += replaySpeedDecrement
		}
		return nil
	}
	if g.level.Completed() {
		return nil
	}
	if v != nil && g.level.CanMoveDown() {
		command.Bus.Do(command.NewMoveDown())
	}
	return nil
}

func (g *game) arrowLeftHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() {
		if g.canSlowDownReplay() {
			g.replaySpeed += replaySpeedDecrement
		}
		return nil
	}
	if g.level.Completed() {
		return nil
	}
	if v != nil && g.level.CanMoveLeft() {
		command.Bus.Do(command.NewMoveLeft())
	}
	return nil
}

func (g *game) undoHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() || g.level.Completed() {
		return nil
	}
	if v != nil {
		command.Bus.UndoLast()
	}
	return nil
}

func (g *game) redoHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() || g.level.Completed() {
		return nil
	}
	if v != nil {
		command.Bus.RedoLast()
	}
	return nil
}

func (g *game) resetHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.level.Completed() {
		return nil
	}
	if v != nil {
		g.reset()
	}
	return nil
}

func (g *game) nextLevelHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() || g.level.Completed() {
		return nil
	}
	if v != nil {
		g.nextLevel()
	}
	return nil
}

func (g *game) previousLevelHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() {
		g.replayPaused = !g.replayPaused
		g.update()
		return nil
	}
	if g.level.Completed() {
		return nil
	}
	if v != nil {
		g.previousLevel()
	}
	return nil
}

func (g *game) replaySolutionHandler(gui *gocui.Gui, v *gocui.View) error {
	if g.replaying() {
		g.replaySpeed = 0
		g.update()
		return nil
	}

	g.reset()
	g.replaySpeed = defaultReplaySpeed

	go func() {
		for g.replaying() && g.replayIndex < len(g.level.Solution) {
			if !g.replayPaused {
				switch g.level.Solution[g.replayIndex] {
				case gokoban.Up:
					command.Bus.Do(command.NewMoveUp())
				case gokoban.Right:
					command.Bus.Do(command.NewMoveRight())
				case gokoban.Down:
					command.Bus.Do(command.NewMoveDown())
				case gokoban.Left:
					command.Bus.Do(command.NewMoveLeft())
				}
				g.replayIndex++
				g.refresh()
			}
			time.Sleep(time.Duration(g.replaySpeed) * time.Millisecond)
		}
	}()

	return nil
}

func (g *game) refreshHandler(gui *gocui.Gui) error {
	return nil
}

func (g *game) quitHandler(gui *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (g *game) keyBindings() error {
	if err := g.gui.SetKeybinding(g.view, gocui.KeyArrowUp, gocui.ModNone, g.arrowUpHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyArrowRight, gocui.ModNone, g.arrowRightHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyArrowDown, gocui.ModNone, g.arrowDownHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyArrowLeft, gocui.ModNone, g.arrowLeftHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyCtrlZ, gocui.ModNone, g.undoHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyCtrlY, gocui.ModNone, g.redoHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyCtrlU, gocui.ModNone, g.undoHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyCtrlR, gocui.ModNone, g.redoHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyCtrlSpace, gocui.ModNone, g.resetHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyCtrlN, gocui.ModNone, g.nextLevelHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyCtrlP, gocui.ModNone, g.previousLevelHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding(g.view, gocui.KeyCtrlS, gocui.ModNone, g.replaySolutionHandler); err != nil {
		return err
	}
	if err := g.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, g.quitHandler); err != nil {
		return err
	}

	return nil
}

func (g *game) movePlayer(course gokoban.Course) {
	g.level.Move(course)

	if g.level.Completed() {
		go func() {
			if g.replaySpeed == 0 {
				g.level.PrintSolution(fmt.Sprintf("my-solution%d.txt", g.lvl))
			}
			time.Sleep(2 * time.Second)
			g.nextLevel()
			g.refresh()
		}()
	}
}

func (g *game) hasPreviousLevel() bool {
	return g.lvl > 1
}

func (g *game) hasNextLevel() bool {
	return g.lvl < g.maxLevel()
}

func (g *game) nextLevel() {
	if g.hasNextLevel() {
		g.lvl++
		g.loadLevel()
	} else {
		g.reset()
	}
}

func (g *game) previousLevel() {
	if g.hasPreviousLevel() {
		g.lvl--
		g.loadLevel()
	}
}

func (g *game) undoLastMove() {
	if g.level.MoveCount() > 0 {
		g.level.UndoLastMove()
		g.update()
	}
}

func (g *game) reset() {
	g.replaySpeed = 0
	g.replayPaused = false
	g.replayIndex = 0
	g.level.Reset()
	g.update()
}

func (g *game) update() {
	v := g.gui.CurrentView()
	if v == nil {
		return
	}
	v.Clear()
	g.print(v)
	g.refresh()
}

func (g *game) refresh() {
	g.gui.Execute(g.refreshHandler)
}

func (g *game) printOption(option, description string, view *gocui.View) {
	_, _ = fmt.Fprintf(view, "%s%s%s%s", bgWhite, black, option, reset)
	_, _ = fmt.Fprintf(view, fmt.Sprintf(" %s ", description))
}

func (g *game) print(view *gocui.View) {
	level := gokoban.Indent(g.level.String(), 40)
	levelInfo := fmt.Sprintf("Level %d/%d", g.lvl, g.maxLevel())
	vw, _ := view.Size()
	s := fmt.Sprintf("%s\n\n%s", gokoban.Indent(levelInfo, (vw-len(levelInfo))/2), level)
	for i := range s {
		symbol := string(s[i])
		if symbol == gokoban.BrickSymbol {
			_, _ = fmt.Fprintf(view, "%s %s", brickColor, reset)
		} else if symbol == gokoban.TargetSymbol {
			_, _ = fmt.Fprintf(view, "%sO%s", targetColor, reset)
		} else if symbol == gokoban.BoxSymbol {
			_, _ = fmt.Fprintf(view, "%s %s", boxColor, reset)
		} else if symbol == gokoban.PlayerSymbol {
			_, _ = fmt.Fprintf(view, "%s %s", playerColor, reset)
		} else {
			_, _ = fmt.Fprint(view, symbol)
		}
	}
	_, _ = fmt.Fprintln(view)
	_, _ = fmt.Fprintln(view)
	_, _ = fmt.Fprint(view, " ")
	if g.replaying() {
		g.printOption("^SPACE", "reset", view)
		if g.canSpeedUpReplay() {
			g.printOption("left", "faster", view)
		}
		if g.canSlowDownReplay() {
			g.printOption("right", "slower", view)
		}
		if g.canSpeedUpReplay() {
			g.printOption("up", "faster", view)
		}
		if g.canSlowDownReplay() {
			g.printOption("down", "slower", view)
		}
		p := "pause"
		if g.replayPaused {
			p = "continue"
		}
		g.printOption("^p", p, view)
		g.printOption("^s", "stop", view)
		g.printOption("^c", "exit", view)

		return
	}
	g.printOption("^SPACE", "reset", view)
	if g.hasPreviousLevel() {
		g.printOption("^p", "previous", view)
	}
	if g.hasNextLevel() {
		g.printOption("^n", "next", view)
	}
	g.printOption("^u", "undo", view)
	g.printOption("^r", "redo", view)
	g.printOption("^z", "undo", view)
	g.printOption("^y", "redo", view)
	g.printOption("^s", "solution", view)
	g.printOption("^c", "exit", view)
}

func (g *game) layout(gui *gocui.Gui) error {
	w := g.level.Width() + 80
	h := g.level.Height() + 3
	maxX, maxY := gui.Size()
	if _, err := gui.SetView(g.view, (maxX-w)/2, (maxY-h)/2-6, (maxX+w)/2, (maxY+h)/2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if err := gui.SetCurrentView(g.view); err != nil {
			return err
		}
	}
	return nil
}
