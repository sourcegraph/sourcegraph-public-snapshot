import { of } from 'rxjs'
import { mergeMap } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import { throttleTimeWindow } from './throttleTimeWindow'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('throttleTimeWindow', () => {
    test('emit the first value (immediately) and last value (at the end) in each time window', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-a-x-y----b---x-cx---|')
            const subs = '^--------------------!'
            const expected = '-a----y---b----xc----(x|)'

            const result = e1.pipe(throttleTimeWindow(5, 1))

            expectObservable(result).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('emit the first 2 values (immediately) and last value (at the end) in each time window', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-a-x-y----b---x-cx---|')
            const subs = '^--------------------!'
            const expected = '-a-x--y---b---x-cx---|'

            const result = e1.pipe(throttleTimeWindow(5, 2))

            expectObservable(result).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('emit the first 3 values (immediately) and last value (at the end) in each time window', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-a-x-y----bxy-z-cx---|')
            const subs = '^--------------------!'
            const expected = '-a-x-y----bxy--zcx---|'

            const result = e1.pipe(throttleTimeWindow(5, 3))

            expectObservable(result).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('simply mirror the source if values are not emitted often enough', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-a--------b-----c----|')
            const subs = '^--------------------!'
            const expected = '-a--------b-----c----|'
            expectObservable(e1.pipe(throttleTimeWindow(5, 2))).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('handle a busy producer emitting a regular repeating sequence', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('abcdefabcdefabcdefabcdefa|')
            const subs = '^------------------------!'
            const expected = 'ab---fab---fab---fab---fa|'

            expectObservable(e1.pipe(throttleTimeWindow(5, 2))).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('complete when source does not emit', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-----|')
            const subs = '^----!'
            const expected = '-----|'

            expectObservable(e1.pipe(throttleTimeWindow(5, 2))).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('raise error when source does not emit and raises error', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-----#')
            const subs = '^----!'
            const expected = '-----#'

            expectObservable(e1.pipe(throttleTimeWindow(1, 2))).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('handle an empty source', () => {
        scheduler().run(({ cold, expectObservable, expectSubscriptions }) => {
            const e1 = cold('|')
            const subs = '(^!)'
            const expected = '|'

            expectObservable(e1.pipe(throttleTimeWindow(3, 2))).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('handle a never source', () => {
        scheduler().run(({ cold, expectObservable, expectSubscriptions }) => {
            const e1 = cold('-')
            const subs = '^'
            const expected = '-'

            expectObservable(e1.pipe(throttleTimeWindow(3, 2))).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('handle a throw source', () => {
        scheduler().run(({ cold, expectObservable, expectSubscriptions }) => {
            const e1 = cold('#')
            const subs = '(^!)'
            const expected = '#'

            expectObservable(e1.pipe(throttleTimeWindow(3, 2))).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('throttle and does not complete when source does not completes', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-a--(bc)-------d----------------')
            const unsub = '-------------------------------!'
            const subs = '^------------------------------!'
            const expected = '-a----c--------d----------------'

            expectObservable(e1.pipe(throttleTimeWindow(5)), unsub).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('not break unsubscription chains when result is unsubscribed explicitly', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-a--(bc)-------d----------------')
            const subs = '^------------------------------!'
            const expected = '-a----c--------d----------------'
            const unsub = '-------------------------------!'

            const result = e1.pipe(
                mergeMap((x: string) => of(x)),
                throttleTimeWindow(5),
                mergeMap((x: string) => of(x))
            )

            expectObservable(result, unsub).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })

    test('throttle values until source raises error', () => {
        scheduler().run(({ hot, expectObservable, expectSubscriptions }) => {
            const e1 = hot('-a--(bc)-------d---------------#')
            const subs = '^------------------------------!'
            const expected = '-a----c--------d---------------#'

            expectObservable(e1.pipe(throttleTimeWindow(5))).toBe(expected)
            expectSubscriptions(e1.subscriptions).toBe(subs)
        })
    })
})
