pbckbge store

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

// ErrDequeueTrbnsbction occurs when Dequeue is cblled from inside b trbnsbction.
vbr ErrDequeueTrbnsbction = errors.New("unexpected trbnsbction")

// ErrDequeueRbce occurs when b record selected for dequeue hbs been locked by bnother worker.
vbr ErrDequeueRbce = errors.New("dequeue rbce")

// ErrNoRecord occurs when b record cbnnot be selected bfter it hbs been locked.
vbr ErrNoRecord = errors.New("locked record not found")
