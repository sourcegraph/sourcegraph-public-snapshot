package web

/*
This file implements a fast router by encoding a list of routes first into a
pseudo-trie, then encoding that pseudo-trie into a state machine realized as
a routing bytecode.

The most interesting part of this router is not its speed (it is quite fast),
but the guarantees it provides. In a naive router, routes are examined one after
another until a match is found, and this is the programming model we want to
support. For any given request ("GET /hello/carl"), there is a list of
"plausible" routes: routes which match the method ("GET"), and which have a
prefix that is a prefix of the requested path ("/" and "/hello/", for instance,
but not "/foobar"). Patterns also have some amount of arbitrary code associated
with them, which tells us whether or not the route matched. Just like the naive
router, our goal is to call each plausible pattern, in the order they were
added, until we find one that matches. The "fast" part here is being smart about
which non-plausible routes we can skip.

First, we sort routes using a pairwise comparison function: sorting occurs as
normal on the prefixes, with the caveat that a route may not be moved past a
route that might also match the same string. Among other things, this means
we're forced to use particularly dumb sorting algorithms, but it only has to
happen once, and there probably aren't even that many routes to begin with. This
logic appears inline in the router's handle() function.

We then build a pseudo-trie from the sorted list of routes. It's not quite a
normal trie because there are certain routes we cannot reorder around other
routes (since we're providing identical semantics to the naive router), but it's
close enough and the basic idea is the same.

Finally, we lower this psuedo-trie from its tree representation to a state
machine bytecode. The bytecode is pretty simple: it contains up to three bytes,
a choice of a bunch of flags, and an index. The state machine is pretty simple:
if the bytes match the next few bytes after the cursor, the instruction matches,
and the state machine advances to the next instruction. If it does not match, it
jumps to the instruction at the index. Various flags modify this basic behavior,
the documentation for which can be found below.

The thing we're optimizing for here over pretty much everything else is memory
locality. We make an effort to lay out both the trie child selection logic and
the matching of long strings consecutively in memory, making both operations
very cheap. In fact, our matching logic isn't particularly asymptotically good,
but in practice the benefits of memory locality outweigh just about everything
else.

Unfortunately, the code implementing all of this is pretty bad (both inefficient
and hard to read). Maybe someday I'll come and take a second pass at it.
*/
type state struct {
	mode smMode
	bs   [3]byte
	i    int32
}
type stateMachine []state

type smMode uint8

// Many combinations of smModes don't make sense, but since this is interal to
// the library I don't feel like documenting them.
const (
	// The two low bits of the mode are used as a length of how many bytes
	// of bs are used. If the length is 0, the node is treated as a
	// wildcard.
	smLengthMask smMode = 3
)

const (
	// Jump to the given index on a match. Ordinarily, the state machine
	// will jump to the state given by the index if the characters do not
	// match.
	smJumpOnMatch smMode = 4 << iota
	// The index is the index of a route to try. If running the route fails,
	// the state machine advances by one.
	smRoute
	// Reset the state machine's cursor into the input string to the state's
	// index value.
	smSetCursor
	// If this bit is set, the machine transitions into a non-accepting
	// state if it matches.
	smFail
)

type trie struct {
	prefix   string
	children []trieSegment
}

// A trie segment is a route matching this point (or -1), combined with a list
// of trie children that follow that route.
type trieSegment struct {
	route    int
	children []trie
}

func buildTrie(routes []route, dp, dr int) trie {
	var t trie
	ts := trieSegment{-1, nil}
	for i, r := range routes {
		if len(r.prefix) != dp {
			continue
		}

		if i == 0 {
			ts.route = 0
		} else {
			subroutes := routes[ts.route+1 : i]
			ts.children = buildTrieSegment(subroutes, dp, dr+ts.route+1)
			t.children = append(t.children, ts)
			ts = trieSegment{i, nil}
		}
	}

	// This could be a little DRYer...
	subroutes := routes[ts.route+1:]
	ts.children = buildTrieSegment(subroutes, dp, dr+ts.route+1)
	t.children = append(t.children, ts)

	for i := range t.children {
		if t.children[i].route != -1 {
			t.children[i].route += dr
		}
	}

	return t
}

func commonPrefix(s1, s2 string) string {
	if len(s1) > len(s2) {
		return commonPrefix(s2, s1)
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return s1[:i]
		}
	}
	return s1
}

func buildTrieSegment(routes []route, dp, dr int) []trie {
	if len(routes) == 0 {
		return nil
	}
	var tries []trie

	start := 0
	p := routes[0].prefix[dp:]
	for i := 1; i < len(routes); i++ {
		ip := routes[i].prefix[dp:]
		cp := commonPrefix(p, ip)
		if len(cp) == 0 {
			t := buildTrie(routes[start:i], dp+len(p), dr+start)
			t.prefix = p
			tries = append(tries, t)
			start = i
			p = ip
		} else {
			p = cp
		}
	}

	t := buildTrie(routes[start:], dp+len(p), dr+start)
	t.prefix = p
	return append(tries, t)
}

// This is a bit confusing, since the encode method on a trie deals exclusively
// with trieSegments (i.e., its children), and vice versa.
//
// These methods are also hideously inefficient, both in terms of memory usage
// and algorithmic complexity. If it ever becomes a problem, maybe we can do
// something smarter than stupid O(N^2) appends, but to be honest, I bet N is
// small (it almost always is :P) and we only do it once at boot anyways.

func (t trie) encode(dp, off int) stateMachine {
	ms := make([]stateMachine, len(t.children))
	subs := make([]stateMachine, len(t.children))
	var l, msl, subl int

	for i, ts := range t.children {
		ms[i], subs[i] = ts.encode(dp, 0)
		msl += len(ms[i])
		l += len(ms[i]) + len(subs[i])
	}

	l++

	m := make(stateMachine, 0, l)
	for i, mm := range ms {
		for j := range mm {
			if mm[j].mode&(smRoute|smSetCursor) != 0 {
				continue
			}

			mm[j].i += int32(off + msl + subl + 1)
		}
		m = append(m, mm...)
		subl += len(subs[i])
	}

	m = append(m, state{mode: smJumpOnMatch, i: -1})

	msl = 0
	for i, sub := range subs {
		msl += len(ms[i])
		for j := range sub {
			if sub[j].mode&(smRoute|smSetCursor) != 0 {
				continue
			}
			if sub[j].i == -1 {
				sub[j].i = int32(off + msl)
			} else {
				sub[j].i += int32(off + len(m))
			}
		}
		m = append(m, sub...)
	}

	return m
}

func (ts trieSegment) encode(dp, off int) (me stateMachine, sub stateMachine) {
	o := 1
	if ts.route != -1 {
		o++
	}
	me = make(stateMachine, len(ts.children)+o)

	me[0] = state{mode: smSetCursor, i: int32(dp)}
	if ts.route != -1 {
		me[1] = state{mode: smRoute, i: int32(ts.route)}
	}

	for i, t := range ts.children {
		p := t.prefix

		bc := copy(me[i+o].bs[:], p)
		me[i+o].mode = smMode(bc) | smJumpOnMatch
		me[i+o].i = int32(off + len(sub))

		for len(p) > bc {
			var bs [3]byte
			p = p[bc:]
			bc = copy(bs[:], p)
			sub = append(sub, state{bs: bs, mode: smMode(bc), i: -1})
		}

		sub = append(sub, t.encode(dp+len(t.prefix), off+len(sub))...)
	}
	return
}

func compile(routes []route) stateMachine {
	if len(routes) == 0 {
		return nil
	}
	t := buildTrie(routes, 0, 0)
	m := t.encode(0, 0)
	for i := range m {
		if m[i].i == -1 {
			m[i].mode = m[i].mode | smFail
		}
	}
	return m
}
