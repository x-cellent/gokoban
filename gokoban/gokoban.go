package gokoban

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
)

type Course int

const (
	Up Course = iota
	Right
	Down
	Left
)

func (c Course) String() string {
	switch c {
	case Up:
		return "u"
	case Right:
		return "r"
	case Down:
		return "d"
	default:
		return "l"
	}
}

func parseCourse(b byte) (Course, error) {
	switch b {
	case 'u':
		return Up, nil
	case 'r':
		return Right, nil
	case 'd':
		return Down, nil
	case 'l':
		return Left, nil
	default:
		return Left, fmt.Errorf("unknown course: %s", string(b))
	}
}

const (
	BrickSymbol          = "#"
	TargetSymbol         = "."
	BoxSymbol            = "$"
	BoxOnTargetSymbol    = "*"
	PlayerSymbol         = "@"
	PlayerOnTargetSymbol = "+"
	FreeSymbol           = " "
)

type fieldKind int

const (
	brick fieldKind = iota
	free
	box
	target
	boxOnTarget
	playerOnTarget
	player
)

func (k fieldKind) isTarget() bool {
	return k == target || k == boxOnTarget || k == playerOnTarget
}

type field struct {
	kind fieldKind
	curr fieldKind
	pos  int
}

func (f *field) currSymbol() string {
	switch f.curr {
	case target:
		return TargetSymbol
	case boxOnTarget:
		return BoxSymbol
	case box:
		return BoxSymbol
	case playerOnTarget:
		return PlayerSymbol
	case player:
		return PlayerSymbol
	case brick:
		return BrickSymbol
	default:
		return FreeSymbol
	}
}

func newField(kind string, pos int) *field {
	var k fieldKind
	switch kind {
	case TargetSymbol:
		k = target
	case BoxSymbol:
		k = box
	case BoxOnTargetSymbol:
		k = boxOnTarget
	case PlayerOnTargetSymbol:
		k = playerOnTarget
	case PlayerSymbol:
		k = player
	case BrickSymbol:
		k = brick
	default:
		k = free
	}
	f := &field{
		kind: k,
		pos:  pos,
	}
	f.reset()
	return f
}

func (f *field) reset() {
	switch f.kind {
	case target:
		f.curr = target
	case box:
		f.curr = box
	case boxOnTarget:
		f.curr = box
	case playerOnTarget:
		f.curr = player
	case player:
		f.curr = player
	case brick:
		f.curr = brick
	default:
		f.curr = free
	}
}

func (f *field) row(width int) int {
	return f.pos / width
}

func (f *field) col(width int) int {
	return f.pos % width
}

type move struct {
	course   Course
	movedBox bool
}

type Level struct {
	width    int
	height   int
	fields   map[int]map[int]*field
	pc       int
	pr       int
	moves    []*move
	Solution []Course
}

func NewLevel(filename, solution string) *Level {
	bb, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(bb), "\n")
	maxWidth := 0
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if len(strings.TrimSpace(line)) == 0 {
			lines = append(lines[:i], lines[i+1:]...)
			continue
		}
		if maxWidth < len(line) {
			maxWidth = len(line)
		}
	}

	level := &Level{
		width:  maxWidth,
		height: len(lines),
		fields: make(map[int]map[int]*field),
	}

	for r, line := range lines {
		for c := 0; c < level.width; c++ {
			kind := " "
			if c < len(line) {
				kind = string(line[c])
			}
			_, ok := level.fields[c]
			if !ok {
				level.fields[c] = make(map[int]*field)
			}
			f := newField(kind, level.pos(c, r))
			level.fields[c][r] = f
			if f.kind == player || f.kind == playerOnTarget {
				level.pc = c
				level.pr = r
			}
		}
	}

	if !level.isValid() {
		panic(fmt.Errorf("level %q is not valid", filename))
	}

	bb, err = ioutil.ReadFile(solution)
	if err != nil {
		panic(fmt.Errorf("solution %q is not valid: %v", solution, err))
	}
	for _, b := range bb {
		if b == '\r' || b == '\n' {
			continue
		}
		c, err := parseCourse(b)
		if err != nil {
			panic(fmt.Errorf("solution %q is not valid: %v", solution, err))
		}
		level.Solution = append(level.Solution, c)
	}

	return level
}

func (l *Level) pos(col, row int) int {
	return row*l.width + col
}

func (l *Level) isValid() bool {
	targetCnt := 0
	boxCnt := 0
	playerCnt := 0

	for _, rows := range l.fields {
		for _, f := range rows {
			switch f.kind {
			case target:
				targetCnt++
			case box:
				boxCnt++
			case playerOnTarget:
				targetCnt++
				playerCnt++
			case player:
				playerCnt++
			}
		}
	}

	if boxCnt == 0 || boxCnt != targetCnt || playerCnt != 1 {
		return false
	}

	for c := 0; c < l.width; c++ {
		// first non-empty field from above must be a brick
		for r := 0; r < l.height; r++ {
			f, ok := l.fields[c][r]
			if !ok || f.kind == free {
				continue
			}
			if f.kind == brick {
				break
			}
			return false
		}
		// first non-empty field from below must be a brick
		for r := l.height - 1; r >= 0; r-- {
			f, ok := l.fields[c][r]
			if !ok || f.kind == free {
				continue
			}
			if f.kind == brick {
				break
			}
			return false
		}
	}
	for r := 0; r < l.height; r++ {
		// first non-empty field from left must be a brick
		for c := 0; c < l.width; c++ {
			f, ok := l.fields[c][r]
			if !ok || f.kind == free {
				continue
			}
			if f.kind == brick {
				break
			}
			return false
		}
		// first non-empty field from right must be a brick
		for c := l.width - 1; c >= 0; c-- {
			f, ok := l.fields[c][r]
			if !ok || f.kind == free {
				continue
			}
			if f.kind == brick {
				break
			}
			return false
		}
	}

	return true
}

func (l *Level) CanMove(course Course) bool {
	switch course {
	case Up:
		return l.CanMoveUp()
	case Right:
		return l.CanMoveRight()
	case Down:
		return l.CanMoveDown()
	case Left:
		return l.CanMoveLeft()
	default:
		return false
	}
}

func (l *Level) CanMoveUp() bool {
	return l.isValidMove(Up)
}

func (l *Level) CanMoveRight() bool {
	return l.isValidMove(Right)
}

func (l *Level) CanMoveDown() bool {
	return l.isValidMove(Down)
}

func (l *Level) CanMoveLeft() bool {
	return l.isValidMove(Left)
}

func (l *Level) Move(course Course) {
	switch course {
	case Up:
		l.MoveUp()
	case Right:
		l.MoveRight()
	case Down:
		l.MoveDown()
	case Left:
		l.MoveLeft()
	}
}

func (l *Level) MoveUp() {
	l.move(Up)
}

func (l *Level) MoveRight() {
	l.move(Right)
}

func (l *Level) MoveDown() {
	l.move(Down)
}

func (l *Level) MoveLeft() {
	l.move(Left)
}

func (l *Level) isValidMove(course Course) bool {
	rc, rr := getRelativeMovement(course)

	tc := l.pc + rc
	tr := l.pr + rr

	to, ok := l.fields[tc][tr]
	if !ok || to.kind == brick {
		return false
	}

	if to.curr == box {
		behindTarget := l.fields[tc+rc][tr+rr]
		if behindTarget.curr == box || behindTarget.kind == brick {
			return false
		}
	}

	return true
}

func (l *Level) move(course Course) {
	rc, rr := getRelativeMovement(course)

	tc := l.pc + rc
	tr := l.pr + rr
	to := l.fields[tc][tr]

	movedBox := to.curr == box
	if movedBox {
		l.fields[tc+rc][tr+rr].curr = box
	}

	to.curr = player

	from := l.fields[l.pc][l.pr]
	if from.kind.isTarget() {
		from.curr = target
	} else {
		from.curr = free
	}

	l.pc += rc
	l.pr += rr

	l.moves = append(l.moves, &move{
		course:   course,
		movedBox: movedBox,
	})
}

func (l *Level) UndoLastMove() {
	if len(l.moves) == 0 {
		return
	}

	var m *move
	m, l.moves = l.moves[len(l.moves)-1], l.moves[:len(l.moves)-1]

	rc, rr := getRelativeMovement(m.course)

	from := l.fields[l.pc][l.pr]
	behindFrom := l.fields[l.pc+rc][l.pr+rr]

	if behindFrom.curr == box && m.movedBox {
		from.curr = box
		if behindFrom.kind.isTarget() {
			behindFrom.curr = target
		} else {
			behindFrom.curr = free
		}
	} else if from.kind.isTarget() {
		from.curr = target
	} else {
		from.curr = free
	}

	var c Course
	switch m.course {
	case Up:
		c = Down
	case Right:
		c = Left
	case Down:
		c = Up
	default:
		c = Right
	}

	rc, rr = getRelativeMovement(c)

	l.pc += rc
	l.pr += rr

	l.fields[l.pc][l.pr].curr = player
}

func (l *Level) Completed() bool {
	for _, rows := range l.fields {
		for _, f := range rows {
			if f.curr != box && f.kind.isTarget() {
				return false
			}
		}
	}

	return true
}

func (l *Level) MoveCount() int {
	return len(l.moves)
}

func (l *Level) Moves() string {
	s := ""
	for _, m := range l.moves {
		s += m.course.String()
	}
	return s
}

func (l *Level) Width() int {
	return l.width
}

func (l *Level) Height() int {
	return l.height
}

func (l *Level) PlayerPosition() (int, int) {
	return l.pc, l.pr
}

func (l *Level) Reset() {
	for c, rows := range l.fields {
		for r, f := range rows {
			if f.kind == player || f.kind == playerOnTarget {
				l.pc = c
				l.pr = r
			}
			f.reset()
		}
	}
	l.moves = l.moves[:0]
}

func (l *Level) PrintSolution(filename string) {
	if !l.Completed() {
		return
	}
	_ = ioutil.WriteFile(filename, []byte(l.Moves()), 0644)
}

func (l *Level) String() string {
	var ff []*field
	for _, rows := range l.fields {
		for _, f := range rows {
			ff = append(ff, f)
		}
	}

	sort.Slice(ff, func(i, j int) bool {
		return ff[i].pos < ff[j].pos
	})

	s := ""
	for i, f := range ff {
		if i > 0 && f.col(l.width) == 0 {
			s += "\n"
		}
		s += f.currSymbol()
	}

	moves := len(l.moves)
	currMoves := fmt.Sprintf("curr: %d moves", moves)
	bestMoves := fmt.Sprintf("best: %d moves", len(l.Solution))
	ident := (l.width - len(bestMoves)) / 2
	s += fmt.Sprintf("\n\n%s\n%s\n", Indent(currMoves, ident), Indent(bestMoves, ident))

	return s
}

func getRelativeMovement(course Course) (int, int) {
	switch course {
	case Up:
		return 0, -1
	case Right:
		return 1, 0
	case Down:
		return 0, 1
	case Left:
		return -1, 0
	default:
		return 0, 0
	}
}

func Indent(s string, indent int) string {
	ss := strings.Split(s, "\n")
	ind := ""
	for i := 0; i < indent; i++ {
		ind += " "
	}
	for i, s := range ss {
		ss[i] = fmt.Sprintf("%s%s", ind, s)
	}
	return strings.Join(ss, "\n")
}
