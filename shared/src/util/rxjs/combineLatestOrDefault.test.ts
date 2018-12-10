import assert from 'assert'
import { of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { combineLatestOrDefault } from './combineLatestOrDefault'

const scheduler = () => new TestScheduler((a, b) => assert.deepStrictEqual(a, b))

describe('combineLatestOrDefault', () => {
    describe('with 0 source observables', () => {
        it('emits an empty array and completes', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(combineLatestOrDefault([], 'x')).toBe('(a|)', {
                    a: [],
                })
            ))
    })

    describe('with 1 source observable', () => {
        it('waits to emit/complete until the source observable emits/completes', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(combineLatestOrDefault([cold('-b-|', { b: 1 })])).toBe('-b-|', {
                    b: [1],
                })
            ))
    })

    describe('of()', () => {
        it('handles 1 of() source observable', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(combineLatestOrDefault([of(1)])).toBe('(a|)', {
                    a: [1],
                })
            ))

        it('handles 2 of() source observables', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(combineLatestOrDefault([of(1), of(2)])).toBe('(a|)', {
                    a: [1, 2],
                })
            ))
    })

    it('handles source observables with staggered emissions', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(combineLatestOrDefault([cold('-b-|', { b: 1 }), cold('--c|', { c: 2 })])).toBe('-bc|', {
                b: [1],
                c: [1, 2],
            })
        ))

    it('handles source observables with staggered completions', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(combineLatestOrDefault([cold('a|', { a: 1 }), cold('a-|', { a: 2 })])).toBe('a-|', {
                a: [1, 2],
            })
        ))

    it('handles source observables with concurrent emissions', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(combineLatestOrDefault([cold('-b-|', { b: 1 }), cold('-b-|', { b: 2 })])).toBe('-b-|', {
                b: [1, 2],
            })
        ))

    it('handles observables with staggered and concurrent emissions', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                combineLatestOrDefault([
                    cold('a-|', { a: 1 }),
                    cold('ab|', { a: 2, b: 3 }),
                    cold('-bc-|', { b: 4, c: 5 }),
                ])
            ).toBe('abc-|', {
                a: [1, 2],
                b: [1, 3, 4],
                c: [1, 3, 5],
            })
        ))

    it('fills in the default value if provided', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                combineLatestOrDefault([cold('a--|', { a: 1 }), cold('-b-|', { b: 2 }), cold('--c|', { c: 3 })], 42)
            ).toBe('abc|', {
                a: [1, 42, 42],
                b: [1, 2, 42],
                c: [1, 2, 3],
            })
        ))
})
