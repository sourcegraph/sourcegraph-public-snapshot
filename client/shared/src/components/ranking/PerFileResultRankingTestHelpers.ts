import type { MatchItem } from './PerFileResultRanking'

// Real match data from searching a file for `error` with results returned in order of line number.
export const testDataRealMatchesByLineNumber: MatchItem[] = [
    {
        highlightRanges: [{ startLine: 0, startCharacter: 51, endLine: 0, endCharacter: 56 }],
        content: '// Package errwrap implements methods to formalize error wrapping in Go.',
        startLine: 0,
        endLine: 0,
    },
    {
        highlightRanges: [{ startLine: 2, startCharacter: 48, endLine: 2, endCharacter: 53 }],
        content: '// All of the top-level functions that take an `error` are built to be able',
        startLine: 2,
        endLine: 2,
    },
    {
        highlightRanges: [
            { startLine: 3, startCharacter: 15, endLine: 3, endCharacter: 20 },
            { startLine: 3, startCharacter: 39, endLine: 3, endCharacter: 44 },
        ],
        content: '// to take any error, not just wrapped errors. This allows you to use errwrap',
        startLine: 3,
        endLine: 3,
    },
    {
        highlightRanges: [{ startLine: 8, startCharacter: 2, endLine: 8, endCharacter: 7 }],
        content: '\t"errors"',
        startLine: 8,
        endLine: 8,
    },
    {
        highlightRanges: [{ startLine: 14, startCharacter: 19, endLine: 14, endCharacter: 24 }],
        content: 'type WalkFunc func(error)',
        startLine: 14,
        endLine: 14,
    },
    {
        highlightRanges: [{ startLine: 20, startCharacter: 11, endLine: 20, endCharacter: 16 }],
        content: '// wrapped error in addition to the wrapper itself. Since all the top-level',
        startLine: 20,
        endLine: 20,
    },
    {
        highlightRanges: [
            { startLine: 24, startCharacter: 8, endLine: 24, endCharacter: 13 },
            { startLine: 24, startCharacter: 19, endLine: 24, endCharacter: 24 },
        ],
        content: '\tWrappedErrors() []error',
        startLine: 24,
        endLine: 24,
    },
    {
        highlightRanges: [{ startLine: 27, startCharacter: 53, endLine: 27, endCharacter: 58 }],
        content: '// Wrap defines that outer wraps inner, returning an error type that',
        startLine: 27,
        endLine: 27,
    },
    {
        highlightRanges: [{ startLine: 28, startCharacter: 3, endLine: 28, endCharacter: 8 }],
        content: '// error be cleanly used with the other methods in this package, such as',
        startLine: 28,
        endLine: 28,
    },
    {
        highlightRanges: [{ startLine: 29, startCharacter: 13, endLine: 29, endCharacter: 18 }],
        content: '// Contains, error, etc.',
        startLine: 29,
        endLine: 29,
    },
    {
        highlightRanges: [{ startLine: 30, startCharacter: 2, endLine: 30, endCharacter: 7 }],
        content: '//error',
        startLine: 30,
        endLine: 30,
    },
    {
        highlightRanges: [
            { startLine: 31, startCharacter: 8, endLine: 31, endCharacter: 13 },
            { startLine: 31, startCharacter: 31, endLine: 31, endCharacter: 36 },
        ],
        content: "// This error won't modify the error message at all (the outer message",
        startLine: 31,
        endLine: 31,
    },
    {
        highlightRanges: [{ startLine: 32, startCharacter: 11, endLine: 32, endCharacter: 37 }],
        content: '// will be error).',
        startLine: 32,
        endLine: 32,
    },
    {
        highlightRanges: [
            { startLine: 33, startCharacter: 23, endLine: 33, endCharacter: 28 },
            { startLine: 33, startCharacter: 30, endLine: 33, endCharacter: 35 },
        ],
        content: 'func Wrap(outer, inner error) error {',
        startLine: 33,
        endLine: 33,
    },
    {
        highlightRanges: [{ startLine: 34, startCharacter: 16, endLine: 34, endCharacter: 21 }],
        content: '\treturn &wrappedError{',
        startLine: 34,
        endLine: 34,
    },
    {
        highlightRanges: [{ startLine: 35, startCharacter: 2, endLine: 35, endCharacter: 7 }],
        content: '\t\terror: outer,',
        startLine: 35,
        endLine: 35,
    },
    {
        highlightRanges: [{ startLine: 40, startCharacter: 18, endLine: 40, endCharacter: 23 }],
        content: '// Wrapf wraps an error with a formatting message. This is similar to using',
        startLine: 40,
        endLine: 40,
    },
    {
        highlightRanges: [
            { startLine: 41, startCharacter: 8, endLine: 41, endCharacter: 13 },
            { startLine: 41, startCharacter: 27, endLine: 41, endCharacter: 32 },
            { startLine: 41, startCharacter: 55, endLine: 41, endCharacter: 60 },
        ],
        content: "// `fmt.Errorf` to wrap an error. If you're using `fmt.Errorf` to wrap",
        startLine: 41,
        endLine: 41,
    },
    {
        highlightRanges: [{ startLine: 42, startCharacter: 3, endLine: 42, endCharacter: 8 }],
        content: '// errors, you should replace it with this.',
        startLine: 42,
        endLine: 42,
    },
    {
        highlightRanges: [{ startLine: 44, startCharacter: 31, endLine: 44, endCharacter: 36 }],
        content: "// format is the format of the error message. The string '{{err}}' will",
        startLine: 44,
        endLine: 44,
    },
    {
        highlightRanges: [{ startLine: 45, startCharacter: 33, endLine: 45, endCharacter: 38 }],
        content: '// be replaced with the original error message.',
        startLine: 45,
        endLine: 45,
    },
    {
        highlightRanges: [{ startLine: 47, startCharacter: 23, endLine: 47, endCharacter: 28 }],
        content: '// Deprecated: Use fmt.Errorf()',
        startLine: 47,
        endLine: 47,
    },
    {
        highlightRanges: [
            { startLine: 48, startCharacter: 30, endLine: 48, endCharacter: 35 },
            { startLine: 48, startCharacter: 37, endLine: 48, endCharacter: 42 },
        ],
        content: 'func Wrapf(format string, err error) error {',
        startLine: 48,
        endLine: 48,
    },
    {
        highlightRanges: [{ startLine: 51, startCharacter: 17, endLine: 51, endCharacter: 22 }],
        content: '\t\touterMsg = err.Error()',
        startLine: 51,
        endLine: 51,
    },
    {
        highlightRanges: [{ startLine: 54, startCharacter: 10, endLine: 54, endCharacter: 15 }],
        content: '\touter := errors.New(strings.Replace(',
        startLine: 54,
        endLine: 54,
    },
    {
        highlightRanges: [
            { startLine: 60, startCharacter: 32, endLine: 60, endCharacter: 37 },
            { startLine: 60, startCharacter: 50, endLine: 60, endCharacter: 55 },
        ],
        content: '// Contains checks if the given error contains an error with the',
        startLine: 60,
        endLine: 60,
    },
    {
        highlightRanges: [{ startLine: 61, startCharacter: 40, endLine: 61, endCharacter: 45 }],
        content: '// message msg. If err is not a wrapped error, this will always return',
        startLine: 61,
        endLine: 61,
    },
    {
        highlightRanges: [{ startLine: 62, startCharacter: 20, endLine: 62, endCharacter: 25 }],
        content: '// false unless the error itself happens to match this msg.',
        startLine: 62,
        endLine: 62,
    },
    {
        highlightRanges: [{ startLine: 63, startCharacter: 18, endLine: 63, endCharacter: 23 }],
        content: 'func Contains(err error, msg string) bool {',
        startLine: 63,
        endLine: 63,
    },
    {
        highlightRanges: [
            { startLine: 67, startCharacter: 36, endLine: 67, endCharacter: 41 },
            { startLine: 67, startCharacter: 54, endLine: 67, endCharacter: 59 },
        ],
        content: '// ContainsType checks if the given error contains an error with',
        startLine: 67,
        endLine: 67,
    },
    {
        highlightRanges: [{ startLine: 68, startCharacter: 56, endLine: 68, endCharacter: 61 }],
        content: '// the same concrete type as v. If err is not a wrapped error, this will',
        startLine: 68,
        endLine: 68,
    },
    {
        highlightRanges: [{ startLine: 70, startCharacter: 22, endLine: 70, endCharacter: 75 }],
        content: 'func ContainsType(err error, v interface{}) bool {',
        startLine: 70,
        endLine: 70,
    },
    {
        highlightRanges: [{ startLine: 74, startCharacter: 62, endLine: 74, endCharacter: 67 }],
        content: '// Get is the same as GetAll but returns the deepest matching error.',
        startLine: 74,
        endLine: 74,
    },
    {
        highlightRanges: [
            { startLine: 75, startCharacter: 13, endLine: 75, endCharacter: 18 },
            { startLine: 75, startCharacter: 32, endLine: 75, endCharacter: 37 },
        ],
        content: 'func Get(err error, msg string) error {',
        startLine: 75,
        endLine: 75,
    },
    {
        highlightRanges: [{ startLine: 84, startCharacter: 70, endLine: 84, endCharacter: 75 }],
        content: '// GetType is the same as GetAllType but returns the deepest matching error.',
        startLine: 84,
        endLine: 84,
    },
    {
        highlightRanges: [
            { startLine: 85, startCharacter: 17, endLine: 85, endCharacter: 22 },
            { startLine: 85, startCharacter: 39, endLine: 85, endCharacter: 44 },
        ],
        content: 'func GetType(err error, v interface{}) error {',
        startLine: 85,
        endLine: 85,
    },
    {
        highlightRanges: [{ startLine: 94, startCharacter: 23, endLine: 94, endCharacter: 28 }],
        content: '// GetAll gets all the errors that might be wrapped in err with the',
        startLine: 94,
        endLine: 94,
    },
    {
        highlightRanges: [{ startLine: 95, startCharacter: 35, endLine: 95, endCharacter: 40 }],
        content: '// given message. The order of the errors is such that the outermost',
        startLine: 95,
        endLine: 95,
    },
    {
        highlightRanges: [{ startLine: 96, startCharacter: 12, endLine: 96, endCharacter: 101 }],
        content: '// matching error (the most recent wrap) is index zero, and so on.',
        startLine: 96,
        endLine: 96,
    },
    {
        highlightRanges: [
            { startLine: 97, startCharacter: 16, endLine: 97, endCharacter: 21 },
            { startLine: 97, startCharacter: 37, endLine: 97, endCharacter: 42 },
        ],
        content: 'func GetAll(err error, msg string) []error {',
        startLine: 97,
        endLine: 97,
    },
    {
        highlightRanges: [{ startLine: 98, startCharacter: 14, endLine: 98, endCharacter: 19 }],
        content: '\tvar result []error',
        startLine: 98,
        endLine: 98,
    },
    {
        highlightRanges: [{ startLine: 100, startCharacter: 20, endLine: 100, endCharacter: 25 }],
        content: '\tWalk(err, func(err error) {',
        startLine: 100,
        endLine: 100,
    },
    {
        highlightRanges: [{ startLine: 101, startCharacter: 9, endLine: 101, endCharacter: 14 }],
        content: '\t\tif err.Error() == msg {',
        startLine: 101,
        endLine: 101,
    },
    {
        highlightRanges: [{ startLine: 109, startCharacter: 27, endLine: 109, endCharacter: 32 }],
        content: '// GetAllType gets all the errors that are the same type as v.',
        startLine: 109,
        endLine: 109,
    },
    {
        highlightRanges: [
            { startLine: 112, startCharacter: 20, endLine: 112, endCharacter: 25 },
            { startLine: 112, startCharacter: 44, endLine: 112, endCharacter: 49 },
        ],
        content: 'func GetAllType(err error, v interface{}) []error {',
        startLine: 112,
        endLine: 112,
    },
    {
        highlightRanges: [{ startLine: 113, startCharacter: 14, endLine: 113, endCharacter: 19 }],
        content: '\tvar result []error',
        startLine: 113,
        endLine: 113,
    },
    {
        highlightRanges: [{ startLine: 119, startCharacter: 20, endLine: 119, endCharacter: 25 }],
        content: '\tWalk(err, func(err error) {',
        startLine: 119,
        endLine: 119,
    },
    {
        highlightRanges: [{ startLine: 133, startCharacter: 30, endLine: 133, endCharacter: 35 }],
        content: '// Walk walks all the wrapped errors in err and calls the callback. If',
        startLine: 133,
        endLine: 133,
    },
    {
        highlightRanges: [{ startLine: 134, startCharacter: 23, endLine: 134, endCharacter: 28 }],
        content: "// err isn't a wrapped error, this will be called once for err. If err",
        startLine: 134,
        endLine: 134,
    },
    {
        highlightRanges: [{ startLine: 135, startCharacter: 16, endLine: 135, endCharacter: 21 }],
        content: '// is a wrapped error, the callback will be called for both the wrapper',
        startLine: 135,
        endLine: 135,
    },
    {
        highlightRanges: [
            { startLine: 136, startCharacter: 19, endLine: 136, endCharacter: 24 },
            { startLine: 136, startCharacter: 48, endLine: 136, endCharacter: 53 },
        ],
        content: '// that implements error as well as the wrapped error itself.',
        startLine: 136,
        endLine: 136,
    },
    {
        highlightRanges: [{ startLine: 137, startCharacter: 14, endLine: 137, endCharacter: 19 }],
        content: 'func Walk(err error, cb WalkFunc) {',
        startLine: 137,
        endLine: 137,
    },
    {
        highlightRanges: [{ startLine: 143, startCharacter: 14, endLine: 143, endCharacter: 19 }],
        content: '\tcase *wrappedError:',
        startLine: 143,
        endLine: 143,
    },
    {
        highlightRanges: [{ startLine: 149, startCharacter: 31, endLine: 149, endCharacter: 36 }],
        content: '\t\tfor _, err := range e.WrappedErrors() {',
        startLine: 149,
        endLine: 149,
    },
    {
        highlightRanges: [{ startLine: 152, startCharacter: 26, endLine: 152, endCharacter: 31 }],
        content: '\tcase interface{ Unwrap() error }:',
        startLine: 152,
        endLine: 152,
    },
    {
        highlightRanges: [
            { startLine: 160, startCharacter: 10, endLine: 160, endCharacter: 15 },
            { startLine: 160, startCharacter: 40, endLine: 160, endCharacter: 45 },
        ],
        content: '// wrappedError is an implementation of error that has both the',
        startLine: 160,
        endLine: 160,
    },
    {
        highlightRanges: [{ startLine: 161, startCharacter: 19, endLine: 161, endCharacter: 24 }],
        content: '// outer and inner errors.',
        startLine: 161,
        endLine: 161,
    },
    {
        highlightRanges: [{ startLine: 162, startCharacter: 12, endLine: 162, endCharacter: 17 }],
        content: 'type wrappedError struct {',
        startLine: 162,
        endLine: 162,
    },
    {
        highlightRanges: [{ startLine: 163, startCharacter: 7, endLine: 163, endCharacter: 12 }],
        content: '\tOuter error',
        startLine: 163,
        endLine: 163,
    },
    {
        highlightRanges: [{ startLine: 164, startCharacter: 7, endLine: 164, endCharacter: 12 }],
        content: '\tInner error',
        startLine: 164,
        endLine: 164,
    },
    {
        highlightRanges: [
            { startLine: 167, startCharacter: 16, endLine: 167, endCharacter: 21 },
            { startLine: 167, startCharacter: 23, endLine: 167, endCharacter: 28 },
        ],
        content: 'func (w *wrappedError) Error() string {',
        startLine: 167,
        endLine: 167,
    },
    {
        highlightRanges: [{ startLine: 168, startCharacter: 16, endLine: 168, endCharacter: 21 }],
        content: '\treturn w.Outer.Error()',
        startLine: 168,
        endLine: 168,
    },
    {
        highlightRanges: [
            { startLine: 171, startCharacter: 16, endLine: 171, endCharacter: 21 },
            { startLine: 171, startCharacter: 30, endLine: 171, endCharacter: 35 },
            { startLine: 171, startCharacter: 41, endLine: 171, endCharacter: 46 },
        ],
        content: 'func (w *wrappedError) WrappedErrors() []error {',
        startLine: 171,
        endLine: 171,
    },
    {
        highlightRanges: [{ startLine: 172, startCharacter: 10, endLine: 172, endCharacter: 15 }],
        content: '\treturn []error{w.Outer, w.Inner}',
        startLine: 172,
        endLine: 172,
    },
    {
        highlightRanges: [
            { startLine: 175, startCharacter: 16, endLine: 175, endCharacter: 21 },
            { startLine: 175, startCharacter: 32, endLine: 175, endCharacter: 37 },
        ],
        content: 'func (w *wrappedError) Unwrap() error {',
        startLine: 175,
        endLine: 175,
    },
]

// Real match data from searching a file for `error` with results returned in order of Zoekt relevance ranking.
export const testDataRealMatchesByZoektRanking: MatchItem[] = [
    {
        highlightRanges: [
            {
                startLine: 167,
                startCharacter: 16,
                endLine: 167,
                endCharacter: 21,
            },
            {
                startLine: 167,
                startCharacter: 23,
                endLine: 167,
                endCharacter: 28,
            },
        ],
        content: 'func (w *wrappedError) Error() string {',
        startLine: 167,
        endLine: 167,
    },
    {
        highlightRanges: [
            {
                startLine: 162,
                startCharacter: 12,
                endLine: 162,
                endCharacter: 17,
            },
        ],
        content: 'type wrappedError struct {',
        startLine: 162,
        endLine: 162,
    },
    {
        highlightRanges: [
            {
                startLine: 24,
                startCharacter: 8,
                endLine: 24,
                endCharacter: 13,
            },
            {
                startLine: 24,
                startCharacter: 19,
                endLine: 24,
                endCharacter: 24,
            },
        ],
        content: '\tWrappedErrors() []error',
        startLine: 24,
        endLine: 24,
    },
    {
        highlightRanges: [
            {
                startLine: 171,
                startCharacter: 16,
                endLine: 171,
                endCharacter: 21,
            },
            {
                startLine: 171,
                startCharacter: 30,
                endLine: 171,
                endCharacter: 35,
            },
            {
                startLine: 171,
                startCharacter: 41,
                endLine: 171,
                endCharacter: 46,
            },
        ],
        content: 'func (w *wrappedError) WrappedErrors() []error {',
        startLine: 171,
        endLine: 171,
    },
    {
        highlightRanges: [
            {
                startLine: 0,
                startCharacter: 51,
                endLine: 0,
                endCharacter: 56,
            },
        ],
        content: '// Package errwrap implements methods to formalize error wrapping in Go.',
        startLine: 0,
        endLine: 0,
    },
    {
        highlightRanges: [
            {
                startLine: 2,
                startCharacter: 48,
                endLine: 2,
                endCharacter: 53,
            },
        ],
        content: '// All of the top-level functions that take an `error` are built to be able',
        startLine: 2,
        endLine: 2,
    },
    {
        highlightRanges: [
            {
                startLine: 3,
                startCharacter: 15,
                endLine: 3,
                endCharacter: 20,
            },
            {
                startLine: 3,
                startCharacter: 39,
                endLine: 3,
                endCharacter: 44,
            },
        ],
        content: '// to take any error, not just wrapped errors. This allows you to use errwrap',
        startLine: 3,
        endLine: 3,
    },
    {
        highlightRanges: [
            {
                startLine: 14,
                startCharacter: 19,
                endLine: 14,
                endCharacter: 24,
            },
        ],
        content: 'type WalkFunc func(error)',
        startLine: 14,
        endLine: 14,
    },
    {
        highlightRanges: [
            {
                startLine: 20,
                startCharacter: 11,
                endLine: 20,
                endCharacter: 16,
            },
        ],
        content: '// wrapped error in addition to the wrapper itself. Since all the top-level',
        startLine: 20,
        endLine: 20,
    },
    {
        highlightRanges: [
            {
                startLine: 27,
                startCharacter: 53,
                endLine: 27,
                endCharacter: 58,
            },
        ],
        content: '// Wrap defines that outer wraps inner, returning an error type that',
        startLine: 27,
        endLine: 27,
    },
    {
        highlightRanges: [
            {
                startLine: 31,
                startCharacter: 34,
                endLine: 31,
                endCharacter: 39,
            },
        ],
        content: "// This function won't modify the error message at all (the outer message",
        startLine: 31,
        endLine: 31,
    },
    {
        highlightRanges: [
            {
                startLine: 33,
                startCharacter: 23,
                endLine: 33,
                endCharacter: 28,
            },
            {
                startLine: 33,
                startCharacter: 30,
                endLine: 33,
                endCharacter: 35,
            },
        ],
        content: 'func Wrap(outer, inner error) error {',
        startLine: 33,
        endLine: 33,
    },
    {
        highlightRanges: [
            {
                startLine: 40,
                startCharacter: 18,
                endLine: 40,
                endCharacter: 23,
            },
        ],
        content: '// Wrapf wraps an error with a formatting message. This is similar to using',
        startLine: 40,
        endLine: 40,
    },
    {
        highlightRanges: [
            {
                startLine: 41,
                startCharacter: 8,
                endLine: 41,
                endCharacter: 13,
            },
            {
                startLine: 41,
                startCharacter: 27,
                endLine: 41,
                endCharacter: 32,
            },
            {
                startLine: 41,
                startCharacter: 55,
                endLine: 41,
                endCharacter: 60,
            },
        ],
        content: "// `fmt.Errorf` to wrap an error. If you're using `fmt.Errorf` to wrap",
        startLine: 41,
        endLine: 41,
    },
    {
        highlightRanges: [
            {
                startLine: 44,
                startCharacter: 31,
                endLine: 44,
                endCharacter: 36,
            },
        ],
        content: "// format is the format of the error message. The string '{{err}}' will",
        startLine: 44,
        endLine: 44,
    },
    {
        highlightRanges: [
            {
                startLine: 45,
                startCharacter: 33,
                endLine: 45,
                endCharacter: 38,
            },
        ],
        content: '// be replaced with the original error message.',
        startLine: 45,
        endLine: 45,
    },
    {
        highlightRanges: [
            {
                startLine: 48,
                startCharacter: 30,
                endLine: 48,
                endCharacter: 35,
            },
            {
                startLine: 48,
                startCharacter: 37,
                endLine: 48,
                endCharacter: 42,
            },
        ],
        content: 'func Wrapf(format string, err error) error {',
        startLine: 48,
        endLine: 48,
    },
    {
        highlightRanges: [
            {
                startLine: 51,
                startCharacter: 17,
                endLine: 51,
                endCharacter: 22,
            },
        ],
        content: '\t\touterMsg = err.Error()',
        startLine: 51,
        endLine: 51,
    },
    {
        highlightRanges: [
            {
                startLine: 60,
                startCharacter: 32,
                endLine: 60,
                endCharacter: 37,
            },
            {
                startLine: 60,
                startCharacter: 50,
                endLine: 60,
                endCharacter: 55,
            },
        ],
        content: '// Contains checks if the given error contains an error with the',
        startLine: 60,
        endLine: 60,
    },
    {
        highlightRanges: [
            {
                startLine: 61,
                startCharacter: 40,
                endLine: 61,
                endCharacter: 45,
            },
        ],
        content: '// message msg. If err is not a wrapped error, this will always return',
        startLine: 61,
        endLine: 61,
    },
    {
        highlightRanges: [
            {
                startLine: 62,
                startCharacter: 20,
                endLine: 62,
                endCharacter: 25,
            },
        ],
        content: '// false unless the error itself happens to match this msg',
        startLine: 62,
        endLine: 62,
    },
    {
        highlightRanges: [
            {
                startLine: 63,
                startCharacter: 18,
                endLine: 63,
                endCharacter: 23,
            },
        ],
        content: 'func Contains(err error, msg string) bool {',
        startLine: 63,
        endLine: 63,
    },
    {
        highlightRanges: [
            {
                startLine: 67,
                startCharacter: 36,
                endLine: 67,
                endCharacter: 41,
            },
            {
                startLine: 67,
                startCharacter: 54,
                endLine: 67,
                endCharacter: 59,
            },
        ],
        content: '// ContainsType checks if the given error contains an error with',
        startLine: 67,
        endLine: 67,
    },
    {
        highlightRanges: [
            {
                startLine: 68,
                startCharacter: 56,
                endLine: 68,
                endCharacter: 61,
            },
        ],
        content: '// the same concrete type as v. If err is not a wrapped error, this will',
        startLine: 68,
        endLine: 68,
    },
    {
        highlightRanges: [
            {
                startLine: 70,
                startCharacter: 22,
                endLine: 70,
                endCharacter: 27,
            },
        ],
        content: 'func ContainsType(err error, v interface{}) bool {',
        startLine: 70,
        endLine: 70,
    },
    {
        highlightRanges: [
            {
                startLine: 74,
                startCharacter: 62,
                endLine: 74,
                endCharacter: 67,
            },
        ],
        content: '// Get is the same as GetAll but returns the deepest matching error.',
        startLine: 74,
        endLine: 74,
    },
    {
        highlightRanges: [
            {
                startLine: 75,
                startCharacter: 13,
                endLine: 75,
                endCharacter: 18,
            },
            {
                startLine: 75,
                startCharacter: 32,
                endLine: 75,
                endCharacter: 37,
            },
        ],
        content: 'func Get(err error, msg string) error {',
        startLine: 75,
        endLine: 75,
    },
    {
        highlightRanges: [
            {
                startLine: 84,
                startCharacter: 70,
                endLine: 84,
                endCharacter: 75,
            },
        ],
        content: '// GetType is the same as GetAllType but returns the deepest matching error.',
        startLine: 84,
        endLine: 84,
    },
    {
        highlightRanges: [
            {
                startLine: 85,
                startCharacter: 17,
                endLine: 85,
                endCharacter: 22,
            },
            {
                startLine: 85,
                startCharacter: 39,
                endLine: 85,
                endCharacter: 44,
            },
        ],
        content: 'func GetType(err error, v interface{}) error {',
        startLine: 85,
        endLine: 85,
    },
    {
        highlightRanges: [
            {
                startLine: 96,
                startCharacter: 12,
                endLine: 96,
                endCharacter: 17,
            },
        ],
        content: '// matching error (the most recent wrap) is index zero, and so on.',
        startLine: 96,
        endLine: 96,
    },
    {
        highlightRanges: [
            {
                startLine: 97,
                startCharacter: 16,
                endLine: 97,
                endCharacter: 21,
            },
            {
                startLine: 97,
                startCharacter: 37,
                endLine: 97,
                endCharacter: 42,
            },
        ],
        content: 'func GetAll(err error, msg string) []error {',
        startLine: 97,
        endLine: 97,
    },
    {
        highlightRanges: [
            {
                startLine: 98,
                startCharacter: 14,
                endLine: 98,
                endCharacter: 19,
            },
        ],
        content: '\tvar result []error',
        startLine: 98,
        endLine: 98,
    },
    {
        highlightRanges: [
            {
                startLine: 100,
                startCharacter: 20,
                endLine: 100,
                endCharacter: 25,
            },
        ],
        content: '\tWalk(err, func(err error) {',
        startLine: 100,
        endLine: 100,
    },
    {
        highlightRanges: [
            {
                startLine: 101,
                startCharacter: 9,
                endLine: 101,
                endCharacter: 14,
            },
        ],
        content: '\t\tif err.Error() == msg {',
        startLine: 101,
        endLine: 101,
    },
    {
        highlightRanges: [
            {
                startLine: 112,
                startCharacter: 20,
                endLine: 112,
                endCharacter: 25,
            },
            {
                startLine: 112,
                startCharacter: 44,
                endLine: 112,
                endCharacter: 49,
            },
        ],
        content: 'func GetAllType(err error, v interface{}) []error {',
        startLine: 112,
        endLine: 112,
    },
    {
        highlightRanges: [
            {
                startLine: 113,
                startCharacter: 14,
                endLine: 113,
                endCharacter: 19,
            },
        ],
        content: '\tvar result []error',
        startLine: 113,
        endLine: 113,
    },
    {
        highlightRanges: [
            {
                startLine: 119,
                startCharacter: 20,
                endLine: 119,
                endCharacter: 25,
            },
        ],
        content: '\tWalk(err, func(err error) {',
        startLine: 119,
        endLine: 119,
    },
    {
        highlightRanges: [
            {
                startLine: 134,
                startCharacter: 23,
                endLine: 134,
                endCharacter: 28,
            },
        ],
        content: "// err isn't a wrapped error, this will be called once for err. If err",
        startLine: 134,
        endLine: 134,
    },
    {
        highlightRanges: [
            {
                startLine: 135,
                startCharacter: 16,
                endLine: 135,
                endCharacter: 21,
            },
        ],
        content: '// is a wrapped error, the callback will be called for both the wrapper',
        startLine: 135,
        endLine: 135,
    },
    {
        highlightRanges: [
            {
                startLine: 136,
                startCharacter: 19,
                endLine: 136,
                endCharacter: 24,
            },
            {
                startLine: 136,
                startCharacter: 48,
                endLine: 136,
                endCharacter: 53,
            },
        ],
        content: '// that implements error as well as the wrapped error itself.',
        startLine: 136,
        endLine: 136,
    },
    {
        highlightRanges: [
            {
                startLine: 137,
                startCharacter: 14,
                endLine: 137,
                endCharacter: 19,
            },
        ],
        content: 'func Walk(err error, cb WalkFunc) {',
        startLine: 137,
        endLine: 137,
    },
    {
        highlightRanges: [
            {
                startLine: 152,
                startCharacter: 26,
                endLine: 152,
                endCharacter: 31,
            },
        ],
        content: '\tcase interface{ Unwrap() error }:',
        startLine: 152,
        endLine: 152,
    },
    {
        highlightRanges: [
            {
                startLine: 160,
                startCharacter: 10,
                endLine: 160,
                endCharacter: 15,
            },
            {
                startLine: 160,
                startCharacter: 40,
                endLine: 160,
                endCharacter: 45,
            },
        ],
        content: '// wrappedError is an implementation of error that has both the',
        startLine: 160,
        endLine: 160,
    },
    {
        highlightRanges: [
            {
                startLine: 163,
                startCharacter: 7,
                endLine: 163,
                endCharacter: 12,
            },
        ],
        content: '\tOuter error',
        startLine: 163,
        endLine: 163,
    },
    {
        highlightRanges: [
            {
                startLine: 164,
                startCharacter: 7,
                endLine: 164,
                endCharacter: 12,
            },
        ],
        content: '\tInner error',
        startLine: 164,
        endLine: 164,
    },
    {
        highlightRanges: [
            {
                startLine: 168,
                startCharacter: 16,
                endLine: 168,
                endCharacter: 21,
            },
        ],
        content: '\treturn w.Outer.Error()',
        startLine: 168,
        endLine: 168,
    },
    {
        highlightRanges: [
            {
                startLine: 172,
                startCharacter: 10,
                endLine: 172,
                endCharacter: 15,
            },
        ],
        content: '\treturn []error{w.Outer, w.Inner}',
        startLine: 172,
        endLine: 172,
    },
    {
        highlightRanges: [
            {
                startLine: 175,
                startCharacter: 16,
                endLine: 175,
                endCharacter: 21,
            },
            {
                startLine: 175,
                startCharacter: 32,
                endLine: 175,
                endCharacter: 37,
            },
        ],
        content: 'func (w *wrappedError) Unwrap() error {',
        startLine: 175,
        endLine: 175,
    },
    {
        highlightRanges: [
            {
                startLine: 8,
                startCharacter: 2,
                endLine: 8,
                endCharacter: 7,
            },
        ],
        content: '\t"errors"',
        startLine: 8,
        endLine: 8,
    },
    {
        highlightRanges: [
            {
                startLine: 34,
                startCharacter: 16,
                endLine: 34,
                endCharacter: 21,
            },
        ],
        content: '\treturn &wrappedError{',
        startLine: 34,
        endLine: 34,
    },
    {
        highlightRanges: [
            {
                startLine: 42,
                startCharacter: 3,
                endLine: 42,
                endCharacter: 8,
            },
        ],
        content: '// errors, you should replace it with this.',
        startLine: 42,
        endLine: 42,
    },
    {
        highlightRanges: [
            {
                startLine: 47,
                startCharacter: 23,
                endLine: 47,
                endCharacter: 28,
            },
        ],
        content: '// Deprecated: Use fmt.Errorf()',
        startLine: 47,
        endLine: 47,
    },
    {
        highlightRanges: [
            {
                startLine: 54,
                startCharacter: 10,
                endLine: 54,
                endCharacter: 15,
            },
        ],
        content: '\touter := errors.New(strings.Replace(',
        startLine: 54,
        endLine: 54,
    },
    {
        highlightRanges: [
            {
                startLine: 94,
                startCharacter: 23,
                endLine: 94,
                endCharacter: 28,
            },
        ],
        content: '// GetAll gets all the errors that might be wrapped in err with the',
        startLine: 94,
        endLine: 94,
    },
    {
        highlightRanges: [
            {
                startLine: 95,
                startCharacter: 35,
                endLine: 95,
                endCharacter: 40,
            },
        ],
        content: '// given message. The order of the errors is such that the outermost',
        startLine: 95,
        endLine: 95,
    },
    {
        highlightRanges: [
            {
                startLine: 109,
                startCharacter: 27,
                endLine: 109,
                endCharacter: 32,
            },
        ],
        content: '// GetAllType gets all the errors that are the same type as v.',
        startLine: 109,
        endLine: 109,
    },
    {
        highlightRanges: [
            {
                startLine: 133,
                startCharacter: 30,
                endLine: 133,
                endCharacter: 35,
            },
        ],
        content: '// Walk walks all the wrapped errors in err and calls the callback. If',
        startLine: 133,
        endLine: 133,
    },
    {
        highlightRanges: [
            {
                startLine: 143,
                startCharacter: 14,
                endLine: 143,
                endCharacter: 19,
            },
        ],
        content: '\tcase *wrappedError:',
        startLine: 143,
        endLine: 143,
    },
    {
        highlightRanges: [
            {
                startLine: 161,
                startCharacter: 19,
                endLine: 161,
                endCharacter: 24,
            },
        ],
        content: '// outer and inner errors.',
        startLine: 161,
        endLine: 161,
    },
    {
        highlightRanges: [
            {
                startLine: 149,
                startCharacter: 31,
                endLine: 149,
                endCharacter: 36,
            },
        ],
        content: '\t\tfor _, err := range e.WrappedErrors() {',
        startLine: 149,
        endLine: 149,
    },
]

// Real match data from searching a file for `if ... {...} patternType:structural`.
export const testDataRealMultilineMatches: MatchItem[] = [
    {
        highlightRanges: [
            {
                startLine: 50,
                startCharacter: 1,
                endLine: 52,
                endCharacter: 2,
            },
        ],
        content: '\tif err != nil {\n\t\touterMsg = err.Error()\n\t}',
        startLine: 50,
        endLine: 52,
    },
    {
        highlightRanges: [
            {
                startLine: 60,
                startCharacter: 19,
                endLine: 65,
                endCharacter: 1,
            },
        ],
        content:
            '// Contains checks if the given error contains an error with the\n// message msg. If err is not a wrapped error, this will always return\n// false unless the error itself happens to match this msg.\nfunc Contains(err error, msg string) bool {\n\treturn len(GetAll(err, msg)) > 0\n}',
        startLine: 60,
        endLine: 65,
    },
    {
        highlightRanges: [
            {
                startLine: 67,
                startCharacter: 23,
                endLine: 72,
                endCharacter: 1,
            },
        ],
        content:
            '// ContainsType checks if the given error contains an error with\n// the same concrete type as v. If err is not a wrapped error, this will\n// check the err itself.\nfunc ContainsType(err error, v interface{}) bool {\n\treturn len(GetAllType(err, v)) > 0\n}',
        startLine: 67,
        endLine: 72,
    },
    {
        highlightRanges: [
            {
                startLine: 77,
                startCharacter: 1,
                endLine: 79,
                endCharacter: 2,
            },
        ],
        content: '\tif len(es) > 0 {\n\t\treturn es[len(es)-1]\n\t}',
        startLine: 77,
        endLine: 79,
    },
    {
        highlightRanges: [
            {
                startLine: 87,
                startCharacter: 1,
                endLine: 89,
                endCharacter: 2,
            },
        ],
        content: '\tif len(es) > 0 {\n\t\treturn es[len(es)-1]\n\t}',
        startLine: 87,
        endLine: 89,
    },
    {
        highlightRanges: [
            {
                startLine: 101,
                startCharacter: 2,
                endLine: 103,
                endCharacter: 3,
            },
        ],
        content: '\t\tif err.Error() == msg {\n\t\t\tresult = append(result, err)\n\t\t}',
        startLine: 101,
        endLine: 103,
    },
    {
        highlightRanges: [
            {
                startLine: 116,
                startCharacter: 1,
                endLine: 118,
                endCharacter: 2,
            },
        ],
        content: '\tif v != nil {\n\t\tsearch = reflect.TypeOf(v).String()\n\t}',
        startLine: 116,
        endLine: 118,
    },
    {
        highlightRanges: [
            {
                startLine: 121,
                startCharacter: 2,
                endLine: 123,
                endCharacter: 3,
            },
        ],
        content: '\t\tif err != nil {\n\t\t\tneedle = reflect.TypeOf(err).String()\n\t\t}',
        startLine: 121,
        endLine: 123,
    },
    {
        highlightRanges: [
            {
                startLine: 125,
                startCharacter: 2,
                endLine: 127,
                endCharacter: 3,
            },
        ],
        content: '\t\tif needle == search {\n\t\t\tresult = append(result, err)\n\t\t}',
        startLine: 125,
        endLine: 127,
    },
]
