import { describe, expect, it } from '@jest/globals'

// Note: This tests the pattern matcher implementation but also acts as a
// verifier that patterns are properly typed against their input value.
// Usage of the @ts-expect-error is intentional and should be kept in place

import { every, matchesValue, some, oneOf, allOf, not, type PatternOfNoInfer } from './patternMatcher'

declare module '@jest/expect' {
    export interface Matchers<R, T> {
        toBeMatchedBy<Data>(pattern: PatternOfNoInfer<T, Data>, expectedData?: Data, initialData?: Data): R
    }
}

expect.extend({
    toBeMatchedBy(actual, pattern, expectedData, initialData) {
        const options = {
            comment: 'Pattern matching',
            isNot: this.isNot,
            promise: this.promise,
        }
        const result = matchesValue(actual, pattern, initialData)
        let dataComparisonResult = false
        if (result.success && expectedData) {
            dataComparisonResult = this.equals(result.data, expectedData)
        }

        return {
            actual,
            message: () =>
                `${this.utils.matcherHint('toMatchPattern', undefined, undefined, options)}\n\n` +
                `Value: ${this.utils.printExpected(actual)}\n` +
                `Pattern: ${this.utils.printExpected(pattern)}\n` +
                `Match: ${this.utils.printReceived(result.success)} (expected: ${this.utils.printExpected(
                    !result.success
                )}))\n` +
                (expectedData
                    ? `Data received: ${this.utils.printReceived(result.success ? result.data : '[no match]')}\n` +
                      `Data expected: ${this.utils.printExpected(expectedData)}\n`
                    : ''),
            pass: result.success && (!expectedData || dataComparisonResult),
        }
    },
})

describe('matchValue', () => {
    describe('base patterns', () => {
        it('allows primtive values as patterns', () => {
            expect(42).toBeMatchedBy(42)
            expect(true).toBeMatchedBy(true)
            expect(false).toBeMatchedBy(false)
            expect('foo').toBeMatchedBy('foo')
            expect('foo').toBeMatchedBy(/^f/)
            expect(null).toBeMatchedBy(null)

            expect(42).not.toBeMatchedBy(21)
            expect(true).not.toBeMatchedBy(false)
            expect(false).not.toBeMatchedBy(true)
            expect('foo').not.toBeMatchedBy('bar')
            // @ts-expect-error cannot use a string to match a number
            expect(42).not.toBeMatchedBy('foo')
            expect(42 as string | number).not.toBeMatchedBy('foo')
        })

        it('allows functions as patterns', () => {
            expect(42).toBeMatchedBy(value => value === 42)
            expect({ field: 42 }).toBeMatchedBy(value => value.field > 0)

            expect(42).not.toBeMatchedBy(value => value < 0)
            expect({ field: 42 }).not.toBeMatchedBy(value => value.field < 0)
            // @ts-expect-error cannot use a number as a pattern for an object
            expect({ field: 42 }).not.toBeMatchedBy(42)
        })

        it('allows objects as patterns', () => {
            expect({ field1: 42, field2: 21 }).toBeMatchedBy({ field1: 42 })
            expect({ field1: 42, field2: 21 }).toBeMatchedBy({ field1: value => value > 0 })
            expect({ field1: 42, field2: 21 }).not.toBeMatchedBy({ field2: 42 })

            // @ts-expect-error cannot use a non-existing field to match an object
            expect({ field1: 42, field2: 21 }).not.toBeMatchedBy({ field3: 42 })
            expect({ field1: undefined } as { field1: undefined | { field2: number } }).not.toBeMatchedBy({
                field1: { field2: 42 },
            })
        })

        it('allows wrapper patterns ', () => {
            expect(42).toBeMatchedBy({ $pattern: 42, $data: {} })
            // Wrapper patterns for objects must use a function as pattern
            expect({ field1: 0 }).toBeMatchedBy({ $pattern: ({ field1 }) => field1 === 0 })
            expect({ field1: 0, field2: 21 }).toBeMatchedBy({ $pattern: ({ field1 }) => field1 === 0, field2: 21 })
            expect({ field1: 0, field2: 21 }).not.toBeMatchedBy({ $pattern: ({ field1 }) => field1 === 0, field2: 42 })
        })
    })

    describe('matching against union types', () => {
        enum TestEnum {
            A = 0,
            B = 1,
            C = 2,
        }
        type TestType = { a: number; b: string } | { a: string }

        it('allows matching against primitve unions', () => {
            expect('foo' as 'foo' | 'bar').toBeMatchedBy('foo')
            expect('foo' as 'foo' | 'bar').not.toBeMatchedBy('bar')

            expect(TestEnum.A as TestEnum).toBeMatchedBy(TestEnum.A)
            expect(TestEnum.A as TestEnum).not.toBeMatchedBy(TestEnum.B)
        })

        it('allows using a property in a pattern that does not exist in all union members', () => {
            expect({ b: 'b' } as TestType).toBeMatchedBy({ b: /^b/ })
            expect({ b: 'b' } as TestType).not.toBeMatchedBy({ a: 'a' })
            expect({ a: 42 } as TestType).not.toBeMatchedBy({ a: 'a' })
            expect({ a: 'a' } as TestType).toBeMatchedBy({ a: 'a' })
        })

        it('properly types function patterns for properties that do not exist in all union members', () => {
            expect({ b: 'a' } as TestType).toBeMatchedBy({ b: v => v?.toUpperCase() === 'A' })
            // @ts-expect-error v is type string|undefined
            expect({ b: 'a' } as TestType).toBeMatchedBy({ b: v => v.toUpperCase() === 'A' })
        })

        it('disallows properties that do not exist on any union member', () => {
            // @ts-expect-error property c doesn't exist
            expect({ a: 'a' } as TestType).not.toBeMatchedBy({ c: 'c' })
        })
    })

    describe('data extraction', () => {
        it('extracts data when there is a match', () => {
            expect(42).toBeMatchedBy({ $pattern: x => x > 0, $data: { larger: true } }, { larger: true })
            expect({
                field1: {
                    field1: 42,
                    field2: [{ field3: 42 }],
                },
            }).toBeMatchedBy(
                { field1: { field1: 42, field2: some({ field3: x => x > 0, $data: { larger: true } }) } },
                { larger: true }
            )
        })

        it('allows functions as data extractor', () => {
            const extractData = (value: number, context: { data: Set<number> }): void => {
                context.data.add(value)
            }
            expect({ a: 42, b: 21 }).toBeMatchedBy(
                {
                    a: { $pattern: 42, $data: extractData },
                    b: { $pattern: 21, $data: extractData },
                },
                new Set([21, 42]),
                new Set()
            )
        })
    })

    describe('some()', () => {
        it('matches against an array of nodes', () => {
            expect({ foo: [42, 21] }).toBeMatchedBy({ foo: some(42) })
            expect({ foo: [42, 21] }).toBeMatchedBy({ foo: some(x => x < 42) })

            expect([42, 21]).not.toBeMatchedBy(some(0))
        })
    })

    describe('every()', () => {
        it('matches against an array of nodes', () => {
            expect({ foo: [42, 21] }).toBeMatchedBy({ foo: every(value => value > 0) })
            expect([42, -1]).not.toBeMatchedBy(every(42))
        })
    })

    describe('oneOf()', () => {
        it('matches one or more patterns against a value', () => {
            expect(42).toBeMatchedBy(oneOf(42, { $pattern: 21, $data: {} }, value => value > 0))

            expect(42).not.toBeMatchedBy(oneOf(21, { $pattern: 21, $data: {} }, value => value < 0))
        })
    })

    describe('allOf()', () => {
        it('matches one ore more patterns against a value', () => {
            expect(42).toBeMatchedBy(allOf(42, { $pattern: 42, $data: {} }, value => value > 0))

            expect(42).not.toBeMatchedBy(allOf(21, { $pattern: 42, $data: {} }, value => value < 0))
        })
    })

    describe('not()', () => {
        it('matches against a single node', () => {
            expect(42).toBeMatchedBy(not(21))
        })
    })
})
