import { MatchItem } from './PerFileResultRanking'

// Real match data from searching a file for `error`.
export const testDataRealMatches: MatchItem[] = [
    {
        highlightRanges: [{ startLine: 0, startCharacter: 51, endLine: 0, endCharacter: 56 }],
        content: '// Package errwrap implements methods to formalize error wrapping in Go.',
        startLine: 0,
    },
    {
        highlightRanges: [{ startLine: 2, startCharacter: 48, endLine: 2, endCharacter: 53 }],
        content: '// All of the top-level functions that take an `error` are built to be able',
        startLine: 2,
    },
    {
        highlightRanges: [{ startLine: 3, startCharacter: 15, endLine: 3, endCharacter: 20 }, { startLine: 3, startCharacter: 39, endLine: 3, endCharacter: 44 }],
        content: '// to take any error, not just wrapped errors. This allows you to use errwrap',
        startLine: 3,
    },
    {
        highlightRanges: [{ startLine: 8, startCharacter: 2, endLine: 8, endCharacter: 7 }],
        content: '\t"errors"',
        startLine: 8,
    },
    {
        highlightRanges: [{ startLine: 14, startCharacter: 19, endLine: 14, endCharacter: 24 }],
        content: 'type WalkFunc func(error)',
        startLine: 14,
    },
    {
        highlightRanges: [{ startLine: 20, startCharacter: 11, endLine: 20, endCharacter: 16 }],
        content: '// wrapped error in addition to the wrapper itself. Since all the top-level',
        startLine: 20,
    },
    {
        highlightRanges: [{ startLine: 24, startCharacter: 8, endLine: 24, endCharacter: 13 }, { startLine: 24, startCharacter: 19, endLine: 24, endCharacter: 24 }],
        content: '\tWrappedErrors() []error',
        startLine: 24,
    },
    {
        highlightRanges: [{ startLine: 27, startCharacter: 53, endLine: 27, endCharacter: 58 }],
        content: '// Wrap defines that outer wraps inner, returning an error type that',
        startLine: 27,
    },
    {
        highlightRanges: [{ startLine: 28, startCharacter: 3, endLine: 28, endCharacter: 8 }],
        content: '// error be cleanly used with the other methods in this package, such as',
        startLine: 28,
    },
    {
        highlightRanges: [{ startLine: 29, startCharacter: 13, endLine: 29, endCharacter: 18 }],
        content: '// Contains, error, etc.',
        startLine: 29,
    },
    {
        highlightRanges: [{ startLine: 30, startCharacter: 2, endLine: 30, endCharacter: 7 }],
        content: '//error',
        startLine: 30,
    },
    {
        highlightRanges: [{ startLine: 31, startCharacter: 8, endLine: 31, endCharacter: 13 }, { startLine: 31, startCharacter: 31, endLine: 31, endCharacter: 36 }],
        content: "// This error won't modify the error message at all (the outer message",
        startLine: 31,
    },
    {
        highlightRanges: [{ startLine: 32, startCharacter: 11, endLine: 32, endCharacter: 37 }],
        content: '// will be error).',
        startLine: 32,
    },
    {
        highlightRanges: [{ startLine: 33, startCharacter: 23, endLine: 33, endCharacter: 28 }, { startLine: 33, startCharacter: 30, endLine: 33, endCharacter: 35 }],
        content: 'func Wrap(outer, inner error) error {',
        startLine: 33,
    },
    {
        highlightRanges: [{ startLine: 34, startCharacter: 16, endLine: 34, endCharacter: 21 }],
        content: '\treturn &wrappedError{',
        startLine: 34,
    },
    {
        highlightRanges: [{ startLine: 35, startCharacter: 2, endLine: 35, endCharacter: 7 }],
        content: '\t\terror: outer,',
        startLine: 35,
    },
    {
        highlightRanges: [{ startLine: 40, startCharacter: 18, endLine: 40, endCharacter: 23 }],
        content: '// Wrapf wraps an error with a formatting message. This is similar to using',
        startLine: 40,
    },
    {
        highlightRanges: [{ startLine: 41, startCharacter: 8, endLine: 41, endCharacter: 13 }, { startLine: 41, startCharacter: 27, endLine: 41, endCharacter: 32 }, { startLine: 41, startCharacter: 55, endLine: 41, endCharacter: 60 }],
        content: "// `fmt.Errorf` to wrap an error. If you're using `fmt.Errorf` to wrap",
        startLine: 41,
    },
    {
        highlightRanges: [{ startLine: 42, startCharacter: 3, endLine: 42, endCharacter: 8 }],
        content: '// errors, you should replace it with this.',
        startLine: 42,
    },
    {
        highlightRanges: [{ startLine: 44, startCharacter: 31, endLine: 44, endCharacter: 36 }],
        content: "// format is the format of the error message. The string '{{err}}' will",
        startLine: 44,
    },
    {
        highlightRanges: [{ startLine: 45, startCharacter: 33, endLine: 45, endCharacter: 38 }],
        content: '// be replaced with the original error message.',
        startLine: 45,
    },
    {
        highlightRanges: [{ startLine: 47, startCharacter: 23, endLine: 47, endCharacter: 28 }],
        content: '// Deprecated: Use fmt.Errorf()',
        startLine: 47,
    },
    {
        highlightRanges: [{ startLine: 48, startCharacter: 30, endLine: 48, endCharacter: 35 }, { startLine: 48, startCharacter: 37, endLine: 48, endCharacter: 42 }],
        content: 'func Wrapf(format string, err error) error {',
        startLine: 48,
    },
    {
        highlightRanges: [{ startLine: 51, startCharacter: 17, endLine: 51, endCharacter: 22 }],
        content: '\t\touterMsg = err.Error()',
        startLine: 51,
    },
    {
        highlightRanges: [{ startLine: 54, startCharacter: 10, endLine: 54, endCharacter: 15 }],
        content: '\touter := errors.New(strings.Replace(',
        startLine: 54,
    },
    {
        highlightRanges: [{ startLine: 60, startCharacter: 32, endLine: 60, endCharacter: 37 }, { startLine: 60, startCharacter: 50, endLine: 60, endCharacter: 55 }],
        content: '// Contains checks if the given error contains an error with the',
        startLine: 60,
    },
    {
        highlightRanges: [{ startLine: 61, startCharacter: 40, endLine: 61, endCharacter: 45 }],
        content: '// message msg. If err is not a wrapped error, this will always return',
        startLine: 61,
    },
    {
        highlightRanges: [{ startLine: 62, startCharacter: 20, endLine: 62, endCharacter: 25 }],
        content: '// false unless the error itself happens to match this msg.',
        startLine: 62,
    },
    {
        highlightRanges: [{ startLine: 63, startCharacter: 18, endLine: 63, endCharacter: 23 }],
        content: 'func Contains(err error, msg string) bool {',
        startLine: 63,
    },
    {
        highlightRanges: [{ startLine: 67, startCharacter: 36, endLine: 67, endCharacter: 41 }, { startLine: 67, startCharacter: 54, endLine: 67, endCharacter: 59 }],
        content: '// ContainsType checks if the given error contains an error with',
        startLine: 67,
    },
    {
        highlightRanges: [{ startLine: 68, startCharacter: 56, endLine: 68, endCharacter: 61 }],
        content: '// the same concrete type as v. If err is not a wrapped error, this will',
        startLine: 68,
    },
    {
        highlightRanges: [{ startLine: 70, startCharacter: 22, endLine: 70, endCharacter: 75 }],
        content: 'func ContainsType(err error, v interface{}) bool {',
        startLine: 70,
    },
    {
        highlightRanges: [{ startLine: 74, startCharacter: 62, endLine: 74, endCharacter: 67 }],
        content: '// Get is the same as GetAll but returns the deepest matching error.',
        startLine: 74,
    },
    {
        highlightRanges: [{ startLine: 75, startCharacter: 13, endLine: 75, endCharacter: 18 }, { startLine: 75, startCharacter: 32, endLine: 75, endCharacter: 37 }],
        content: 'func Get(err error, msg string) error {',
        startLine: 75,
    },
    {
        highlightRanges: [{ startLine: 84, startCharacter: 70, endLine: 84, endCharacter: 75 }],
        content: '// GetType is the same as GetAllType but returns the deepest matching error.',
        startLine: 84,
    },
    {
        highlightRanges: [{ startLine: 85, startCharacter: 17, endLine: 85, endCharacter: 22 }, { startLine: 85, startCharacter: 39, endLine: 85, endCharacter: 44 }],
        content: 'func GetType(err error, v interface{}) error {',
        startLine: 85,
    },
    {
        highlightRanges: [{ startLine: 94, startCharacter: 23, endLine: 94, endCharacter: 28 }],
        content: '// GetAll gets all the errors that might be wrapped in err with the',
        startLine: 94,
    },
    {
        highlightRanges: [{ startLine: 95, startCharacter: 35, endLine: 95, endCharacter: 40 }],
        content: '// given message. The order of the errors is such that the outermost',
        startLine: 95,
    },
    {
        highlightRanges: [{ startLine: 96, startCharacter: 12, endLine: 96, endCharacter: 101 }],
        content: '// matching error (the most recent wrap) is index zero, and so on.',
        startLine: 96,
    },
    {
        highlightRanges: [{ startLine: 97, startCharacter: 16, endLine: 97, endCharacter: 21 }, { startLine: 97, startCharacter: 37, endLine: 97, endCharacter: 42 }],
        content: 'func GetAll(err error, msg string) []error {',
        startLine: 97,
    },
    {
        highlightRanges: [{ startLine: 98, startCharacter: 14, endLine: 98, endCharacter: 19 }],
        content: '\tvar result []error',
        startLine: 98,
    },
    {
        highlightRanges: [{ startLine: 100, startCharacter: 20, endLine: 100, endCharacter: 25 }],
        content: '\tWalk(err, func(err error) {',
        startLine: 100,
    },
    {
        highlightRanges: [{ startLine: 101, startCharacter: 9, endLine: 101, endCharacter: 14 }],
        content: '\t\tif err.Error() == msg {',
        startLine: 101,
    },
    {
        highlightRanges: [{ startLine: 109, startCharacter: 27, endLine: 109, endCharacter: 32 }],
        content: '// GetAllType gets all the errors that are the same type as v.',
        startLine: 109,
    },
    {
        highlightRanges: [{ startLine: 112, startCharacter: 20, endLine: 112, endCharacter: 25 }, { startLine: 112, startCharacter: 44, endLine: 112, endCharacter: 49 }],
        content: 'func GetAllType(err error, v interface{}) []error {',
        startLine: 112,
    },
    {
        highlightRanges: [{ startLine: 113, startCharacter: 14, endLine: 113, endCharacter: 19 }],
        content: '\tvar result []error',
        startLine: 113,
    },
    {
        highlightRanges: [{ startLine: 119, startCharacter: 20, endLine: 119, endCharacter: 25 }],
        content: '\tWalk(err, func(err error) {',
        startLine: 119,
    },
    {
        highlightRanges: [{ startLine: 133, startCharacter: 30, endLine: 133, endCharacter: 35 }],
        content: '// Walk walks all the wrapped errors in err and calls the callback. If',
        startLine: 133,
    },
    {
        highlightRanges: [{ startLine: 134, startCharacter: 23, endLine: 134, endCharacter: 28 }],
        content: "// err isn't a wrapped error, this will be called once for err. If err",
        startLine: 134,
    },
    {
        highlightRanges: [{ startLine: 135, startCharacter: 16, endLine: 135, endCharacter: 21 }],
        content: '// is a wrapped error, the callback will be called for both the wrapper',
        startLine: 135,
    },
    {
        highlightRanges: [{ startLine: 136, startCharacter: 19, endLine: 136, endCharacter: 24 }, { startLine: 136, startCharacter: 48, endLine: 136, endCharacter: 53 }],
        content: '// that implements error as well as the wrapped error itself.',
        startLine: 136,
    },
    {
        highlightRanges: [{ startLine: 137, startCharacter: 14, endLine: 137, endCharacter: 19 }],
        content: 'func Walk(err error, cb WalkFunc) {',
        startLine: 137,
    },
    {
        highlightRanges: [{ startLine: 143, startCharacter: 14, endLine: 143, endCharacter: 19 }],
        content: '\tcase *wrappedError:',
        startLine: 143,
    },
    {
        highlightRanges: [{ startLine: 149, startCharacter: 31, endLine: 149, endCharacter: 36 }],
        content: '\t\tfor _, err := range e.WrappedErrors() {',
        startLine: 149,
    },
    {
        highlightRanges: [{ startLine: 152, startCharacter: 26, endLine: 152, endCharacter: 31 }],
        content: '\tcase interface{ Unwrap() error }:',
        startLine: 152,
    },
    {
        highlightRanges: [{ startLine: 160, startCharacter: 10, endLine: 160, endCharacter: 15 }, { startLine: 160, startCharacter: 40, endLine: 160, endCharacter: 45 }],
        content: '// wrappedError is an implementation of error that has both the',
        startLine: 160,
    },
    {
        highlightRanges: [{ startLine: 161, startCharacter: 19, endLine: 161, endCharacter: 24 }],
        content: '// outer and inner errors.',
        startLine: 161,
    },
    {
        highlightRanges: [{ startLine: 162, startCharacter: 12, endLine: 162, endCharacter: 17 }],
        content: 'type wrappedError struct {',
        startLine: 162,
    },
    {
        highlightRanges: [{ startLine: 163, startCharacter: 7, endLine: 163, endCharacter: 12 }],
        content: '\tOuter error',
        startLine: 163,
    },
    {
        highlightRanges: [{ startLine: 164, startCharacter: 7, endLine: 164, endCharacter: 12 }],
        content: '\tInner error',
        startLine: 164,
    },
    {
        highlightRanges: [{ startLine: 167, startCharacter: 16, endLine: 167, endCharacter: 21 }, { startLine: 167, startCharacter: 23, endLine: 167, endCharacter: 28 }],
        content: 'func (w *wrappedError) Error() string {',
        startLine: 167,
    },
    {
        highlightRanges: [{ startLine: 168, startCharacter: 16, endLine: 168, endCharacter: 21 }],
        content: '\treturn w.Outer.Error()',
        startLine: 168,
    },
    {
        highlightRanges: [{ startLine: 171, startCharacter: 16, endLine: 171, endCharacter: 21 }, { startLine: 171, startCharacter: 30, endLine: 171, endCharacter: 35 }, { startLine: 171, startCharacter: 41, endLine: 171, endCharacter: 46 }],
        content: 'func (w *wrappedError) WrappedErrors() []error {',
        startLine: 171,
    },
    {
        highlightRanges: [{ startLine: 172, startCharacter: 10, endLine: 172, endCharacter: 15 }],
        content: '\treturn []error{w.Outer, w.Inner}',
        startLine: 172,
    },
    {
        highlightRanges: [{ startLine: 175, startCharacter: 16, endLine: 175, endCharacter: 21 }, { startLine: 175, startCharacter: 32, endLine: 175, endCharacter: 37 }],
        content: 'func (w *wrappedError) Unwrap() error {',
        startLine: 175,
    },
]
