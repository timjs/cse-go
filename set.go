/*
Copyright © 2010 John Asmuth. All Rights Reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list
of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this
list of conditions and the following disclaimer in the documentation and/or other
materials provided with the distribution.

3. The name of the author may not be used to endorse or promote products derived
from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY [LICENSOR] "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF
USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

//target:gohash.googlecode.com/hg/hashset

//Hashset backed by a linked list.
package main

import (
	"container/list"
)

type Hasher interface {
	Hashcode() uint64
}
type Equalser interface {
	Equals(other interface{}) bool
}

type HashFunc func(key interface{}) uint64
type EqualsFunc func(a, b interface{}) bool

func MethodHash(key interface{}) uint64 {
	return key.(Hasher).Hashcode()
}
func MethodEquals(a, b interface{}) bool {
	return a.(Equalser).Equals(b)
}

type Set struct {
	bins     map[uint64]*list.List
	hasher   HashFunc
	equalser EqualsFunc
	count    int
}

func NewSet() *Set {
	return NewSetFuncs(MethodHash, MethodEquals)
}

func NewSetFuncs(hasher HashFunc, equalser EqualsFunc) (me *Set) {
	me = new(Set)
	me.bins = make(map[uint64]*list.List)
	me.hasher = hasher
	me.equalser = equalser
	me.count = 0
	return
}

func (me *Set) Keys() (out chan interface{}) {
	out = make(chan interface{})
	go func(out chan interface{}) {
		for _, bin := range me.bins {
			for c := bin.Front(); c != nil; c = c.Next() {
				out <- c.Value
			}
		}
		close(out)
	}(out)
	return
}

func (me *Set) Insert(key interface{}) {
	index := me.hasher(key)
	bin, exists := me.bins[index]
	if !exists {
		bin = list.New()
		me.bins[index] = bin
	}
	for c := bin.Front(); c != nil; c = c.Next() {
		if me.equalser(c.Value, key) {
			c.Value = key
			return
		}
	}
	bin.PushFront(key)
	me.count++
}

func (me *Set) Remove(key interface{}) {
	index := me.hasher(key)
	if bin, exists := me.bins[index]; exists {
		for c := bin.Front(); c != nil; c = c.Next() {
			if me.equalser(c.Value, key) {
				bin.Remove(c)
				me.count--
				return
			}
		}
	}
}

func (me *Set) Get(key interface{}) (item interface{}, ok bool) {
	index := me.hasher(key)
	if bin, exists := me.bins[index]; exists {
		for c := bin.Front(); c != nil; c = c.Next() {
			if me.equalser(c.Value, key) {
				item, ok = c.Value, true
				return
			}
		}
	}
	return
}

func (me *Set) Contains(key interface{}) (exists bool) {
	_, exists = me.Get(key)
	return
}

func (me *Set) Size() int {
	return me.count
}
