package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Expressions

type (
	Expr struct {
		name  Name
		sub   Repl
		left  *Expr
		right *Expr
	}
	Name string
	Repl int
)

// Display

func (expr *Expr) display() string {
	var buffer bytes.Buffer
	if expr.name == "" {
		buffer.WriteString(fmt.Sprint(expr.sub))
	} else {
		buffer.WriteString(string(expr.name))
		if expr.left != nil { // && expr.right != nil {
			buffer.WriteByte('(')
			buffer.WriteString(expr.left.display())
			buffer.WriteByte(',')
			buffer.WriteString(expr.right.display())
			buffer.WriteByte(')')
		}
	}
	return buffer.String()
}

// Parser

func isLetter(b byte) bool { return 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' }

type Parser struct {
	bufio.Reader
}

func newParser(input string) *Parser {
	return &Parser{*bufio.NewReader(strings.NewReader(input))}
}

func (parser *Parser) readWhile(test func(byte) bool) string {
	var buffer bytes.Buffer
	for b, e := parser.ReadByte(); e == nil; b, e = parser.ReadByte() {
		if test(b) {
			buffer.WriteByte(b)
		} else {
			parser.UnreadByte()
			break
		}
	}
	return buffer.String()
}

func (parser *Parser) readName() string {
	return parser.readWhile(isLetter)
}

func (parser *Parser) readVar(name string) *Expr {
	return &Expr{name: Name(name)}
}

func (parser *Parser) readApp(name string) *Expr {
	parser.ReadByte() // '('
	left := parser.readExpr()
	parser.ReadByte() // ','
	right := parser.readExpr()
	parser.ReadByte() // ')'
	return &Expr{name: Name(name), left: left, right: right}
}

func (parser *Parser) readExpr() *Expr {
	name := parser.readName()
	switch b, _ := parser.Peek(1); b[0] {
	case '(':
		return parser.readApp(name)
	default:
		return parser.readVar(name)
	}
}

// Elimination

type State struct {
	dict map[Expr]Repl // ERROR! uses equallity on pointers `left` and `right` in Expr!
	num  Repl
}

func newState() *State {
	return &State{dict: make(map[Expr]Repl), num: 1}
}

func (expr *Expr) cseMut(state *State) {
	if repl := state.dict[*expr]; repl != 0 {
		*expr = Expr{sub: repl}
	} else {
		state.dict[*expr] = state.num
		state.num++
		if expr.left != nil { // && expr.right != nil {
			expr.left.cseMut(state)
			expr.right.cseMut(state)
		}
	}
}

func (expr *Expr) cse(state *State) *Expr {
	if repl := state.dict[*expr]; repl != 0 {
		return &Expr{sub: repl}
	} else {
		state.dict[*expr] = state.num
		state.num++
		if expr.left != nil { // && expr.right != nil {
			l := expr.left.cse(state)
			r := expr.right.cse(state)
			return &Expr{left: l, right: r, name: expr.name}
		} else {
			return &Expr{name: expr.name}
		}
	}
}

// Main

func main() {
	stdin := bufio.NewReader(os.Stdin)
	firstLine, _ := stdin.ReadString('\n')
	firstLine = strings.TrimSpace(firstLine)
	lineCount, _ := strconv.Atoi(firstLine)
	for i := 0; i < lineCount; i++ {
		input, _ := stdin.ReadString('\n')
		parser := newParser(input)
		expr := parser.readExpr()
		result := expr.cse(newState())
		fmt.Println(result.display())
	}
}
