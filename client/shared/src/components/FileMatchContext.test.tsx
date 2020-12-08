import { range } from 'lodash'
import { MatchItem } from './FileMatch'
import { calculateMatchGroups, mergeContext } from './FileMatchContext'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

describe('components/FileMatchContext', () => {
    describe('mergeContext', () => {
        test('handles empty input', () => {
            expect(mergeContext(1, [])).toEqual([])
        })
        test('does not merge context when there is only one line', () => {
            expect(mergeContext(1, [{ line: 5 }])).toEqual([[{ line: 5 }]])
        })
        test('merges overlapping context', () => {
            expect(mergeContext(1, [{ line: 5 }, { line: 6 }])).toEqual([[{ line: 5 }, { line: 6 }]])
        })
        test('merges adjacent context', () => {
            expect(mergeContext(1, [{ line: 5 }, { line: 8 }])).toEqual([[{ line: 5 }, { line: 8 }]])
        })
        test('does not merge context when far enough apart', () => {
            expect(mergeContext(1, [{ line: 5 }, { line: 9 }])).toEqual([[{ line: 5 }], [{ line: 9 }]])
        })
    })

    describe('calculateMatchGroups', () => {
        test('simple', () => {
            const maxMatches = 3
            const context = 1
            const [, grouped] = calculateMatchGroups(testData6ConsecutiveMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "line": 0,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 1,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 2,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 3,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": true
                      }
                    ],
                    "position": {
                      "line": 1,
                      "character": 1
                    },
                    "startLine": 0,
                    "endLine": 4
                  }
                ]
            `)
        })

        test('no context', () => {
            const maxMatches = 3
            const context = 0
            const [, grouped] = calculateMatchGroups(testData6ConsecutiveMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "line": 0,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 1,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 2,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      }
                    ],
                    "position": {
                      "line": 1,
                      "character": 1
                    },
                    "startLine": 0,
                    "endLine": 3
                  }
                ]
            `)
        })

        test('complex grouping', () => {
            const maxMatches = 10
            const context = 2
            const [, grouped] = calculateMatchGroups(testDataRealMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "line": 0,
                        "character": 51,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 2,
                        "character": 48,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 3,
                        "character": 15,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 3,
                        "character": 39,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 8,
                        "character": 2,
                        "highlightLength": 5,
                        "IsInContext": false
                      }
                    ],
                    "position": {
                      "line": 1,
                      "character": 52
                    },
                    "startLine": 0,
                    "endLine": 11
                  },
                  {
                    "matches": [
                      {
                        "line": 14,
                        "character": 19,
                        "highlightLength": 5,
                        "IsInContext": false
                      }
                    ],
                    "position": {
                      "line": 15,
                      "character": 20
                    },
                    "startLine": 12,
                    "endLine": 17
                  },
                  {
                    "matches": [
                      {
                        "line": 20,
                        "character": 11,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 24,
                        "character": 8,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 24,
                        "character": 19,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 27,
                        "character": 53,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 28,
                        "character": 3,
                        "highlightLength": 5,
                        "IsInContext": true
                      },
                      {
                        "line": 29,
                        "character": 13,
                        "highlightLength": 5,
                        "IsInContext": true
                      }
                    ],
                    "position": {
                      "line": 21,
                      "character": 12
                    },
                    "startLine": 18,
                    "endLine": 30
                  }
                ]
            `)
        })
    })
})

// "error" matched 5 times, once per line.
const testData6ConsecutiveMatches: MatchItem[] = range(0, 6).map(index => ({
    highlightRanges: [{ start: 0, highlightLength: 5 }],
    preview: 'error',
    line: index,
}))

// Real match data from searching a file for `error`.
const testDataRealMatches: MatchItem[] = [
    {
        highlightRanges: [{ start: 51, highlightLength: 5 }],
        preview: '// Package errwrap implements methods to formalize error wrapping in Go.',
        line: 0,
    },
    {
        highlightRanges: [{ start: 48, highlightLength: 5 }],
        preview: '// All of the top-level functions that take an `error` are built to be able',
        line: 2,
    },
    {
        highlightRanges: [{ start: 15, highlightLength: 5 }],
        preview: '// to take any error, not just wrapped errors. This allows you to use errwrap',
        line: 3,
    },
    {
        highlightRanges: [{ start: 39, highlightLength: 5 }],
        preview: '// to take any error, not just wrapped errors. This allows you to use errwrap',
        line: 3,
    },
    {
        highlightRanges: [{ start: 2, highlightLength: 5 }],
        preview: '\t"errors"',
        line: 8,
    },
    {
        highlightRanges: [{ start: 19, highlightLength: 5 }],
        preview: 'type WalkFunc func(error)',
        line: 14,
    },
    {
        highlightRanges: [{ start: 11, highlightLength: 5 }],
        preview: '// wrapped error in addition to the wrapper itself. Since all the top-level',
        line: 20,
    },
    {
        highlightRanges: [{ start: 8, highlightLength: 5 }],
        preview: '\tWrappedErrors() []error',
        line: 24,
    },
    {
        highlightRanges: [{ start: 19, highlightLength: 5 }],
        preview: '\tWrappedErrors() []error',
        line: 24,
    },
    {
        highlightRanges: [{ start: 53, highlightLength: 5 }],
        preview: '// Wrap defines that outer wraps inner, returning an error type that',
        line: 27,
    },
    {
        highlightRanges: [{ start: 3, highlightLength: 5 }],
        preview: '// error be cleanly used with the other methods in this package, such as',
        line: 28,
    },
    {
        highlightRanges: [{ start: 13, highlightLength: 5 }],
        preview: '// Contains, error, etc.',
        line: 29,
    },
    {
        highlightRanges: [{ start: 2, highlightLength: 5 }],
        preview: '//error',
        line: 30,
    },
    {
        highlightRanges: [{ start: 8, highlightLength: 5 }],
        preview: "// This error won't modify the error message at all (the outer message",
        line: 31,
    },
    {
        highlightRanges: [{ start: 31, highlightLength: 5 }],
        preview: "// This error won't modify the error message at all (the outer message",
        line: 31,
    },
    {
        highlightRanges: [{ start: 11, highlightLength: 5 }],
        preview: '// will be error).',
        line: 32,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        preview: 'func Wrap(outer, inner error) error {',
        line: 33,
    },
    {
        highlightRanges: [{ start: 30, highlightLength: 5 }],
        preview: 'func Wrap(outer, inner error) error {',
        line: 33,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        preview: '\treturn &wrappedError{',
        line: 34,
    },
    {
        highlightRanges: [{ start: 2, highlightLength: 5 }],
        preview: '\t\terror: outer,',
        line: 35,
    },
    {
        highlightRanges: [{ start: 18, highlightLength: 5 }],
        preview: '// Wrapf wraps an error with a formatting message. This is similar to using',
        line: 40,
    },
    {
        highlightRanges: [{ start: 8, highlightLength: 5 }],
        preview: "// `fmt.Errorf` to wrap an error. If you're using `fmt.Errorf` to wrap",
        line: 41,
    },
    {
        highlightRanges: [{ start: 27, highlightLength: 5 }],
        preview: "// `fmt.Errorf` to wrap an error. If you're using `fmt.Errorf` to wrap",
        line: 41,
    },
    {
        highlightRanges: [{ start: 55, highlightLength: 5 }],
        preview: "// `fmt.Errorf` to wrap an error. If you're using `fmt.Errorf` to wrap",
        line: 41,
    },
    {
        highlightRanges: [{ start: 3, highlightLength: 5 }],
        preview: '// errors, you should replace it with this.',
        line: 42,
    },
    {
        highlightRanges: [{ start: 31, highlightLength: 5 }],
        preview: "// format is the format of the error message. The string '{{err}}' will",
        line: 44,
    },
    {
        highlightRanges: [{ start: 33, highlightLength: 5 }],
        preview: '// be replaced with the original error message.',
        line: 45,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        preview: '// Deprecated: Use fmt.Errorf()',
        line: 47,
    },
    {
        highlightRanges: [{ start: 30, highlightLength: 5 }],
        preview: 'func Wrapf(format string, err error) error {',
        line: 48,
    },
    {
        highlightRanges: [{ start: 37, highlightLength: 5 }],
        preview: 'func Wrapf(format string, err error) error {',
        line: 48,
    },
    {
        highlightRanges: [{ start: 17, highlightLength: 5 }],
        preview: '\t\touterMsg = err.Error()',
        line: 51,
    },
    {
        highlightRanges: [{ start: 10, highlightLength: 5 }],
        preview: '\touter := errors.New(strings.Replace(',
        line: 54,
    },
    {
        highlightRanges: [{ start: 32, highlightLength: 5 }],
        preview: '// Contains checks if the given error contains an error with the',
        line: 60,
    },
    {
        highlightRanges: [{ start: 50, highlightLength: 5 }],
        preview: '// Contains checks if the given error contains an error with the',
        line: 60,
    },
    {
        highlightRanges: [{ start: 40, highlightLength: 5 }],
        preview: '// message msg. If err is not a wrapped error, this will always return',
        line: 61,
    },
    {
        highlightRanges: [{ start: 20, highlightLength: 5 }],
        preview: '// false unless the error itself happens to match this msg.',
        line: 62,
    },
    {
        highlightRanges: [{ start: 18, highlightLength: 5 }],
        preview: 'func Contains(err error, msg string) bool {',
        line: 63,
    },
    {
        highlightRanges: [{ start: 36, highlightLength: 5 }],
        preview: '// ContainsType checks if the given error contains an error with',
        line: 67,
    },
    {
        highlightRanges: [{ start: 54, highlightLength: 5 }],
        preview: '// ContainsType checks if the given error contains an error with',
        line: 67,
    },
    {
        highlightRanges: [{ start: 56, highlightLength: 5 }],
        preview: '// the same concrete type as v. If err is not a wrapped error, this will',
        line: 68,
    },
    {
        highlightRanges: [{ start: 22, highlightLength: 5 }],
        preview: 'func ContainsType(err error, v interface{}) bool {',
        line: 70,
    },
    {
        highlightRanges: [{ start: 62, highlightLength: 5 }],
        preview: '// Get is the same as GetAll but returns the deepest matching error.',
        line: 74,
    },
    {
        highlightRanges: [{ start: 13, highlightLength: 5 }],
        preview: 'func Get(err error, msg string) error {',
        line: 75,
    },
    {
        highlightRanges: [{ start: 32, highlightLength: 5 }],
        preview: 'func Get(err error, msg string) error {',
        line: 75,
    },
    {
        highlightRanges: [{ start: 70, highlightLength: 5 }],
        preview: '// GetType is the same as GetAllType but returns the deepest matching error.',
        line: 84,
    },
    {
        highlightRanges: [{ start: 17, highlightLength: 5 }],
        preview: 'func GetType(err error, v interface{}) error {',
        line: 85,
    },
    {
        highlightRanges: [{ start: 39, highlightLength: 5 }],
        preview: 'func GetType(err error, v interface{}) error {',
        line: 85,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        preview: '// GetAll gets all the errors that might be wrapped in err with the',
        line: 94,
    },
    {
        highlightRanges: [{ start: 35, highlightLength: 5 }],
        preview: '// given message. The order of the errors is such that the outermost',
        line: 95,
    },
    {
        highlightRanges: [{ start: 12, highlightLength: 5 }],
        preview: '// matching error (the most recent wrap) is index zero, and so on.',
        line: 96,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        preview: 'func GetAll(err error, msg string) []error {',
        line: 97,
    },
    {
        highlightRanges: [{ start: 37, highlightLength: 5 }],
        preview: 'func GetAll(err error, msg string) []error {',
        line: 97,
    },
    {
        highlightRanges: [{ start: 14, highlightLength: 5 }],
        preview: '\tvar result []error',
        line: 98,
    },
    {
        highlightRanges: [{ start: 20, highlightLength: 5 }],
        preview: '\tWalk(err, func(err error) {',
        line: 100,
    },
    {
        highlightRanges: [{ start: 9, highlightLength: 5 }],
        preview: '\t\tif err.Error() == msg {',
        line: 101,
    },
    {
        highlightRanges: [{ start: 27, highlightLength: 5 }],
        preview: '// GetAllType gets all the errors that are the same type as v.',
        line: 109,
    },
    {
        highlightRanges: [{ start: 20, highlightLength: 5 }],
        preview: 'func GetAllType(err error, v interface{}) []error {',
        line: 112,
    },
    {
        highlightRanges: [{ start: 44, highlightLength: 5 }],
        preview: 'func GetAllType(err error, v interface{}) []error {',
        line: 112,
    },
    {
        highlightRanges: [{ start: 14, highlightLength: 5 }],
        preview: '\tvar result []error',
        line: 113,
    },
    {
        highlightRanges: [{ start: 20, highlightLength: 5 }],
        preview: '\tWalk(err, func(err error) {',
        line: 119,
    },
    {
        highlightRanges: [{ start: 30, highlightLength: 5 }],
        preview: '// Walk walks all the wrapped errors in err and calls the callback. If',
        line: 133,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        preview: "// err isn't a wrapped error, this will be called once for err. If err",
        line: 134,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        preview: '// is a wrapped error, the callback will be called for both the wrapper',
        line: 135,
    },
    {
        highlightRanges: [{ start: 19, highlightLength: 5 }],
        preview: '// that implements error as well as the wrapped error itself.',
        line: 136,
    },
    {
        highlightRanges: [{ start: 48, highlightLength: 5 }],
        preview: '// that implements error as well as the wrapped error itself.',
        line: 136,
    },
    {
        highlightRanges: [{ start: 14, highlightLength: 5 }],
        preview: 'func Walk(err error, cb WalkFunc) {',
        line: 137,
    },
    {
        highlightRanges: [{ start: 14, highlightLength: 5 }],
        preview: '\tcase *wrappedError:',
        line: 143,
    },
    {
        highlightRanges: [{ start: 31, highlightLength: 5 }],
        preview: '\t\tfor _, err := range e.WrappedErrors() {',
        line: 149,
    },
    {
        highlightRanges: [{ start: 26, highlightLength: 5 }],
        preview: '\tcase interface{ Unwrap() error }:',
        line: 152,
    },
    {
        highlightRanges: [{ start: 10, highlightLength: 5 }],
        preview: '// wrappedError is an implementation of error that has both the',
        line: 160,
    },
    {
        highlightRanges: [{ start: 40, highlightLength: 5 }],
        preview: '// wrappedError is an implementation of error that has both the',
        line: 160,
    },
    {
        highlightRanges: [{ start: 19, highlightLength: 5 }],
        preview: '// outer and inner errors.',
        line: 161,
    },
    {
        highlightRanges: [{ start: 12, highlightLength: 5 }],
        preview: 'type wrappedError struct {',
        line: 162,
    },
    {
        highlightRanges: [{ start: 7, highlightLength: 5 }],
        preview: '\tOuter error',
        line: 163,
    },
    {
        highlightRanges: [{ start: 7, highlightLength: 5 }],
        preview: '\tInner error',
        line: 164,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        preview: 'func (w *wrappedError) Error() string {',
        line: 167,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        preview: 'func (w *wrappedError) Error() string {',
        line: 167,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        preview: '\treturn w.Outer.Error()',
        line: 168,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        preview: 'func (w *wrappedError) WrappedErrors() []error {',
        line: 171,
    },
    {
        highlightRanges: [{ start: 30, highlightLength: 5 }],
        preview: 'func (w *wrappedError) WrappedErrors() []error {',
        line: 171,
    },
    {
        highlightRanges: [{ start: 41, highlightLength: 5 }],
        preview: 'func (w *wrappedError) WrappedErrors() []error {',
        line: 171,
    },
    {
        highlightRanges: [{ start: 10, highlightLength: 5 }],
        preview: '\treturn []error{w.Outer, w.Inner}',
        line: 172,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        preview: 'func (w *wrappedError) Unwrap() error {',
        line: 175,
    },
    {
        highlightRanges: [{ start: 32, highlightLength: 5 }],
        preview: 'func (w *wrappedError) Unwrap() error {',
        line: 175,
    },
]
