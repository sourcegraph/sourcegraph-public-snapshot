import assert from 'assert'

import { describe, it } from '@jest/globals'
import { TestScheduler } from 'rxjs/testing'

import { emitLoading, LOADING, type MaybeLoadingResult } from './loading'

const inputAlphabet: Record<'l' | 'e' | 'i' | 'r', MaybeLoadingResult<number | null>> = {
    // loading
    l: { isLoading: true, result: null },
    // empty
    e: { isLoading: false, result: null },
    // intermediate result
    i: { isLoading: true, result: 1 },
    // result
    r: { isLoading: false, result: 2 },
}

const outputAlphabet = {
    // undefined
    u: undefined,
    // empty
    e: null,
    // loading
    l: LOADING,
    // intermediate result
    i: inputAlphabet.i.result,
    // result
    r: inputAlphabet.r.result,
}

describe('emitLoading()', () => {
    it('emits an empty result if the source emits an empty result before the loader delay', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('l 10ms e', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 10ms e', outputAlphabet)
        })
    })
    it('emits a loader if the source has not emitted after the loader delay', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('400ms r', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 299ms l 99ms r', outputAlphabet)
        })
    })
    it('emits an empty result if the source completes without a result', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('400ms |', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 299ms l 99ms (e|)', outputAlphabet)
        })
    })
    it('emits an empty result if the source completes without a result before the loader delay', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('10ms |', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 9ms (e|)', outputAlphabet)
        })
    })
    it('emits an empty result if the source completes after an empty loading result', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('l 400ms |', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 299ms l 100ms (e|)', outputAlphabet)
        })
    })
    it('emits the last result if the source completes after an intermediate non-empty loading result', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('-i-|', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('ui-(i|)', outputAlphabet)
        })
    })
    it('errors if the source errors', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const error = new Error('test')
            const source = cold('10ms #', inputAlphabet, error)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 9ms #', outputAlphabet, error)
        })
    })
    it('emits a loader if the source has not emitted a result after the loader delay', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('l 400ms r', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 299ms l 100ms r', outputAlphabet)
        })
    })
    it('emits a loader if the source first emits an empty result, but then starts loading again', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('e 10ms l 400ms r', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('(ue) 296ms l 111ms r', outputAlphabet)
        })
    })
    it('emits a loader if the source first emits an empty result, but then starts loading again after the loader delay already passed', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('e 400ms l 400ms r', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('(ue) 397ms l 400ms r', outputAlphabet)
        })
    })
    it('hides the loader when the source emits an empty result', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('l 400ms e', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 299ms l 100ms e', outputAlphabet)
        })
    })
    it('emits intermediate results before the loader delay', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('l 10ms i 10ms r', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 10ms i 10ms r', outputAlphabet)
        })
    })
    it('emits intermediate results after the loader delay and showing a loader', () => {
        const scheduler = new TestScheduler(assert.deepStrictEqual)
        scheduler.run(({ cold, expectObservable }) => {
            const source = cold('l 400ms i 10ms r', inputAlphabet)
            expectObservable(source.pipe(emitLoading(300, null))).toBe('u 299ms l 100ms i 10ms r', outputAlphabet)
        })
    })
})
