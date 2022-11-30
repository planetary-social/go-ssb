package common

import "github.com/boreq/errors"

// ReceiveLogSequence is zero-indexed. This type has nothing to do with the
// sequence field of Scuttlebutt messages. It is a part of the system which
// simulates the behaviour of go-ssb's receive log.
type ReceiveLogSequence struct {
	seq int
}

func NewReceiveLogSequence(seq int) (ReceiveLogSequence, error) {
	if seq < 0 {
		return ReceiveLogSequence{}, errors.New("sequence can't be negative")
	}

	return ReceiveLogSequence{seq: seq}, nil
}

func MustNewReceiveLogSequence(seq int) ReceiveLogSequence {
	v, err := NewReceiveLogSequence(seq)
	if err != nil {
		panic(err)
	}

	return v
}

func (r ReceiveLogSequence) Int() int {
	return r.seq
}
