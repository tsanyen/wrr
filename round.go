package wrr

import (
	"errors"
	"sync"
)

const (
	DEFAULT_CAP = 128
)

var (
	ErrEmptyNodes = errors.New("empty nodes")
)

type node struct {
	indirect interface{}
	current  int
	weight   int
}

type RoundRobin struct {
	sync.RWMutex
	nodes []node
}

type newOption struct {
	c int
}

type optionOp func(*newOption)

func New(ops ...optionOp) *RoundRobin {
	var c newOption = newOption{
		c: DEFAULT_CAP,
	}
	for _, op := range ops {
		op(&c)
	}
	return &RoundRobin{
		nodes: make([]node, 1, c.c),
	}
}

func (r *RoundRobin) Put(indirect interface{}, weight int) {
	r.Lock()
	defer r.Unlock()
	if indirect == nil {
		panic("can not put nil")
	}
	p := len(r.nodes)
	r.nodes = append(r.nodes, node{
		indirect, 0, weight,
	})
	p >>= 1
	if parent := r.nodes[p]; parent.indirect != nil {
		sibling := parent
		r.nodes = append(r.nodes, sibling) //maybe extending cap
		r.nodes[p] = node{
			nil,
			sibling.current,
			sibling.weight,
		}
		parent = r.nodes[p]
	}
	for ; p > 0; p >>= 1 {
		r.nodes[p].weight += weight
	}
	// r.reset()
}

func (r *RoundRobin) Round() (interface{}, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.nodes) == 1 {
		return nil, ErrEmptyNodes
	}
	return r.round(), nil
}

func (r *RoundRobin) Reset() {
	r.Lock()
	defer r.Unlock()
	r.reset()
}

func (r *RoundRobin) reset() {
	for p, l := 1, len(r.nodes); p < l; p++ {
		r.nodes[p].current = 0
	}
}

func (r *RoundRobin) round() interface{} {
	// if len(r.nodes) == 1 {
	// 	panic("empty nodes")
	// }
	p := 1
	for {
		n := r.nodes[p]
		if n.indirect != nil {
			return n.indirect
		}
		left := &r.nodes[p<<1]
		right := &r.nodes[p<<1+1]
		if left.current+left.weight > right.current+right.weight {
			left.current -= right.weight
			right.current += right.weight
			p <<= 1
		} else {
			right.current -= left.weight
			left.current += left.weight
			p = p<<1 + 1
		}
	}
}

func (r *RoundRobin) Remove(indirect interface{}) {
	r.Lock()
	defer r.Unlock()
	var (
		l = len(r.nodes)
		p = l >> 1
	)
	for ; p < l; p++ {
		if r.nodes[p].indirect == indirect {
			break
		}
	}
	if p == l {
		return
	}
	discarded := r.nodes[p]
	for i := p; i > 0; i >>= 1 {
		r.nodes[i].weight -= discarded.weight
	}
	if p < l-1 { // not last one
		last := r.nodes[l-1]
		for i := l - 1; i > 0; i >>= 1 { //remove last one
			r.nodes[i].weight -= last.weight
		}
		for i := p; i > 0; i >>= 1 { //backfill last one
			r.nodes[i].weight += last.weight
		}
		r.nodes[p] = last
		r.nodes[(l-2)>>1] = r.nodes[l-2]
		r.nodes = r.nodes[:l-2]
	} else {
		r.nodes = r.nodes[:l-1]
	}
	// r.reset()
}

func (r *RoundRobin) Gets(c int) []interface{} {
	r.Lock()
	defer r.Unlock()
	if len(r.nodes) == 1 {
		return []interface{}{}
	}
	if c <= 0 {
		c = r.nodes[1].weight
	}
	var ps = make([]interface{}, c)
	for n, i := len(ps), 0; i < n; i++ {
		ps[i] = r.round()
	}
	return ps
}

func WithCap(c int) optionOp {
	if c <= 0 {
		c = DEFAULT_CAP
	}
	if c&0x1 == 0x1 {
		c = (c >> 0x1) << 0x2
	}
	return func(o *newOption) {
		o.c = c
	}
}
