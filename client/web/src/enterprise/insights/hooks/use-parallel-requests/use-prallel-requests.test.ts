import { renderHook, act } from '@testing-library/react'
import { type Observable, type ObservableInput, of } from 'rxjs'
import { delay, map, switchMap, tap } from 'rxjs/operators'
import sinon from 'sinon'
import { vi, beforeEach, describe, expect, it } from 'vitest'

import { createUseParallelRequestsHook, type FetchResult } from './use-parallel-request'

vi.useFakeTimers()

describe('useParallelRequests', () => {
    let useParallelRequests: <D>(request: () => ObservableInput<D>) => FetchResult<D>

    beforeEach(() => {
        const { query } = createUseParallelRequestsHook({ maxRequests: 1 })

        useParallelRequests = query
    })

    describe('with single request', () => {
        it('should executes immediately without queueing', async () => {
            const request = sinon.spy<() => Promise<{ payload: string }>>(() => Promise.resolve({ payload: 'data' }))
            const requestFactory = () => request()

            const { result } = renderHook(() => useParallelRequests(requestFactory))

            expect(result.current.loading).toBeTruthy()
            expect(result.current.data).toBe(undefined)
            expect(result.current.error).toBe(undefined)

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runAllTimers()
            })

            expect(result.current.loading).toBe(false)
            expect(result.current.data).toStrictEqual({ payload: 'data' })
            expect(result.current.error).toBe(undefined)
        })

        it('should handle error', async () => {
            const networkError = new Error('Network error')
            const request = sinon.spy<() => Promise<unknown>>(() => Promise.reject(networkError))
            const requestFactory = () => request()

            const { result } = renderHook(() => useParallelRequests(requestFactory))

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runAllTimers()
            })

            expect(result.current).toStrictEqual({
                data: undefined,
                loading: false,
                error: networkError,
            })
        })

        it('should cancel promise-like request on unmount', async () => {
            const request = sinon.spy<() => Promise<unknown>>(() => Promise.resolve({}))
            const requestFactory = () => request()

            const { result, unmount } = renderHook(() => useParallelRequests(requestFactory))

            expect(result.current).toStrictEqual({
                data: undefined,
                error: undefined,
                loading: true,
            })

            unmount()

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runAllTimers()
            })

            sinon.assert.notCalled(request)
        })

        it('should cancel stream on unmount', async () => {
            const startStreamCallback = sinon.spy()
            const endStreamCallback = sinon.spy()
            const request = sinon.spy<() => Observable<unknown>>(() =>
                of(null).pipe(
                    tap(startStreamCallback),
                    delay(0),
                    switchMap(() =>
                        of(null).pipe(
                            map(() => ({ data: 'api payload' })),
                            tap(endStreamCallback)
                        )
                    )
                )
            )
            const requestFactory = () => request()

            const { result, unmount } = renderHook(() => useParallelRequests(requestFactory))

            expect(result.current).toStrictEqual({
                data: undefined,
                error: undefined,
                loading: true,
            })

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runOnlyPendingTimers()
            })

            // The First level stream was resolved
            sinon.assert.calledOnce(startStreamCallback)
            sinon.assert.notCalled(endStreamCallback)

            unmount()

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runOnlyPendingTimers()
            })

            // The second level wasn't resolved in fact it was cancelled
            // cause unmount happened
            sinon.assert.calledOnce(startStreamCallback)
            sinon.assert.notCalled(endStreamCallback)
        })
    })

    describe('with two requests', () => {
        it('should execute two requests one by one with queueing', async () => {
            const request1 = sinon.spy<() => Promise<{ payload: string }>>(() => Promise.resolve({ payload: 'data1' }))
            const requestFactory1 = () => request1()
            const request2 = sinon.spy<() => Promise<{ payload: string }>>(() => Promise.resolve({ payload: 'data2' }))
            const requestFactory2 = () => request2()

            const { result: result1 } = renderHook(() => useParallelRequests(requestFactory1))
            const { result: result2 } = renderHook(() => useParallelRequests(requestFactory2))

            expect(result1.current).toStrictEqual({
                data: undefined,
                error: undefined,
                loading: true,
            })

            expect(result2.current).toStrictEqual({
                data: undefined,
                error: undefined,
                loading: true,
            })

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runAllTimers()
            })

            expect(result1.current).toStrictEqual({
                data: { payload: 'data1' },
                error: undefined,
                loading: false,
            })

            expect(result2.current).toStrictEqual({
                data: undefined,
                error: undefined,
                loading: true,
            })

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runAllTimers()
            })

            expect(result2.current).toStrictEqual({
                data: { payload: 'data2' },
                error: undefined,
                loading: false,
            })
        })

        it('should cancel second request if unmount happened', async () => {
            const firstRequest = sinon.spy(() => Promise.resolve({ data: 'payload1' }))
            const firstRequestFactory = () => firstRequest()
            const secondRequest = sinon.spy(() => Promise.resolve({ data: 'payload2' }))
            const secondRequestFactory = () => secondRequest()

            const { result: firstResult } = renderHook(() => useParallelRequests(firstRequestFactory))
            const { result: secondResult, unmount: unmountSecond } = renderHook(() =>
                useParallelRequests(secondRequestFactory)
            )

            expect(firstResult.current).toStrictEqual({
                data: undefined,
                error: undefined,
                loading: true,
            })

            expect(secondResult.current).toStrictEqual({
                data: undefined,
                error: undefined,
                loading: true,
            })

            unmountSecond()

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runAllTimers()
            })

            sinon.assert.calledOnce(firstRequest)
            sinon.assert.notCalled(secondRequest)

            // eslint-disable-next-line @typescript-eslint/require-await
            await act(async () => {
                vi.runAllTimers()
            })

            sinon.assert.notCalled(secondRequest)
        })
    })
})
