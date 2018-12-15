import * as assert from 'assert'
import { Subject } from 'rxjs'
import { isPromise, isSubscribable, tryCatchPromise } from './util'

describe('tryCatchPromise', () => {
    it('returns a resolved promise with the synchronous result', async () =>
        assert.strictEqual(await tryCatchPromise(() => 1), 1))

    it('returns a resolved promise with the asynchronous result', async () =>
        assert.strictEqual(await tryCatchPromise(() => Promise.resolve(1)), 1))

    const ERROR = new Error('x')
    it('returns a rejected promise with the synchronous exception', () => {
        const p = tryCatchPromise(() => {
            throw ERROR
        })
        let rejected: any
        return p.then(undefined, v => (rejected = v)).then(() => {
            assert.strictEqual(rejected, ERROR)
        })
    })

    it('returns a rejected promise with the asynchronous error', () => {
        const p = tryCatchPromise(
            () => Promise.reject(ERROR) // tslint:disable-line:no-floating-promises
        )
        let rejected: any
        return p.then(undefined, v => (rejected = v)).then(() => {
            assert.strictEqual(rejected, ERROR)
        })
    })
})

describe('isPromise', () => {
    it('returns true for promises', () =>
        assert.strictEqual(
            isPromise(
                new Promise<any>(() => {
                    /* noop */
                })
            ),
            true
        ))
    it('returns true for non-promises', () => {
        assert.strictEqual(isPromise(1), false)
        assert.strictEqual(isPromise({ then: 1 }), false)
        assert.strictEqual(isPromise(new Subject<any>()), false)
    })
})

describe('isSubscribable', () => {
    it('returns true for subscribables', () => assert.strictEqual(isSubscribable(new Subject<any>()), true))
    it('returns true for non-subscribables', () => {
        assert.strictEqual(isSubscribable(1), false)
        assert.strictEqual(isSubscribable({ subscribe: 1 }), false)
        assert.strictEqual(
            isSubscribable(
                new Promise<any>(() => {
                    /* noop */
                })
            ),
            false
        )
    })
})
