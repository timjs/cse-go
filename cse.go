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
		hash  Hash
		name  Name
		sub   Repl
		left  *Expr
		right *Expr
	}
	Name string
	Repl int
	Hash uint64
)

// Hasher
// => Doesn't throw compiler errors if not implemented!

func (expr Expr) Hashcode() uint64 {
	return uint64(expr.hash)
}

func (expr1 Expr) Equals(other interface{}) bool {
	if expr2, ok := other.(Expr); ok { // assert(other: Expr)
		if expr2.left != nil { // && expr2.right != nil {
			// User Equals() to descend tree pointers
			// FIXME something going mad...
			return expr1.name == expr2.name && expr1.left.Equals(expr2.left) && expr1.right.Equals(expr2.right)
		} else {
			return expr1.name == expr2.name && expr1.sub == expr2.sub
		}
	} else {
		return false
	}
}

func (name Name) Hashcode() (hash uint64) {
	for i, c := range name {
		hash += uint64(c) * 2 << uint64(i)
	}
	return
}

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

func (parser *Parser) readName() Name {
	return Name(parser.readWhile(isLetter))
}

func (parser *Parser) readVar(name Name) *Expr {
	hash := name.Hashcode()
	return &Expr{hash: Hash(hash), name: name}
}

func (parser *Parser) readApp(name Name) *Expr {
	parser.ReadByte() // '('
	left := parser.readExpr()
	parser.ReadByte() // ','
	right := parser.readExpr()
	parser.ReadByte() // ')'
	hash := name.Hashcode() + left.Hashcode() + right.Hashcode()
	return &Expr{hash: Hash(hash), name: name, left: left, right: right}
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
	dict *Map
	num  Repl
}

func newState() *State {
	return &State{dict: NewMap(), num: 1}
}

func (expr *Expr) cseMut(state *State) {
	if repl, ok := state.dict.Get(*expr); ok {
		*expr = Expr{sub: repl.(Repl)} // assert(state.dict: map[*Expr]Repl)
	} else {
		state.dict.Put(*expr, state.num)
		state.num++
		if expr.left != nil { // && expr.right != nil {
			expr.left.cseMut(state)
			expr.right.cseMut(state)
		}
	}
}

func (expr *Expr) cse(state *State) *Expr {
	if repl, ok := state.dict.Get(*expr); ok {
		return &Expr{sub: repl.(Repl)} // assert(state.dict: map[*Expr]Repl)
	} else {
		state.dict.Put(*expr, state.num)
		state.num++
		if expr.left != nil { // && expr.right != nil {
			l := expr.left.cse(state)
			r := expr.right.cse(state)
			return &Expr{hash: expr.hash, left: l, right: r, name: expr.name}
		} else {
			return &Expr{hash: expr.hash, name: expr.name}
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
		expr.cseMut(newState())
		fmt.Println(expr.display())
	}
}
