//go:generate peg -inline -switch grammar.peg

// Package naturaldate provides natural date time parsing.
package naturaldate

import (
	"strings"
	"time"
)

// day duration.
var day = time.Hour * 24

// week duration.
var week = time.Hour * 24 * 7

// Direction is the direction used for ambiguous expressions.
type Direction int

// Directions available.
const (
	Past Direction = iota
	Future
)

// Option function.
type Option func(*parser)

// WithDirection sets the direction used for ambiguous expressions. By default
// the Past direction is used, so "sunday" will be the previous Sunday, rather
// than the next Sunday.
func WithDirection(d Direction) Option {
	return func(p *parser) {
		switch d {
		case Past:
			p.direction = -1
		case Future:
			p.direction = 1
		default:
			panic("unhandled direction")
		}
	}
}

// Parse query string.
func Parse(s string, ref time.Time, options ...Option) (time.Time, error) {
	p := &parser{
		Buffer:    strings.ToLower(s),
		direction: -1,
		t:         ref,
	}

	for _, o := range options {
		o(p)
	}

	p.Init()

	if err := p.Parse(); err != nil {
		return time.Time{}, err
	}

	p.Execute()
	// p.PrintSyntaxTree()
	return p.t, nil
}

// withDirection returns duration with direction.
func (p *parser) withDirection(d time.Duration) time.Duration {
	return d * time.Duration(p.direction)
}

// prevWeekday returns the previous week day relative to time t.
func prevWeekday(t time.Time, day time.Weekday) time.Time {
	d := t.Weekday() - day
	if d <= 0 {
		d += 7
	}
	return t.Add(-time.Hour * 24 * time.Duration(d))
}

// nextWeekday returns the next week day relative to time t.
func nextWeekday(t time.Time, day time.Weekday) time.Time {
	d := day - t.Weekday()
	if d <= 0 {
		d += 7
	}
	return t.Add(time.Hour * 24 * time.Duration(d))
}

// nextMonth returns the next month relative to time t.
func nextMonth(t time.Time, month time.Month) time.Time {
	y := t.Year()
	if month-t.Month() <= 0 {
		y++
	}
	_, _, day := t.Date()
	hour, min, sec := t.Clock()
	return time.Date(y, month, day, hour, min, sec, 0, t.Location())
}

// prevMonth returns the next month relative to time t.
func prevMonth(t time.Time, month time.Month) time.Time {
	y := t.Year()
	if t.Month()-month <= 0 {
		y--
	}
	_, _, day := t.Date()
	hour, min, sec := t.Clock()
	return time.Date(y, month, day, hour, min, sec, 0, t.Location())
}

// truncateDay returns a date truncated to the day.
func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
