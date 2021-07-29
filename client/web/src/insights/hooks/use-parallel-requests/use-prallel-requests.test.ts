import { renderHook, act } from '@testing-library/react-hooks'
import { ObservableInput } from 'rxjs'
import sinon from 'sinon'

import { createUseParallelRequestsHook, FetchResult } from './use-parallel-request'

jest.useFakeTimers()

describe('useParallelRequests', () => {
    let useParallelRequests: <D>(request: () => ObservableInput<D>) => FetchResult<D>

    beforeEach(() => {
        useParallelRequests = createUseParallelRequestsHook({ maxRequests: 1 })
    })

    it('should execute single execution immediately without queueing', async () => {
        const request = sinon.spy<() => Promise<{ payload: string }>>(() => Promise.resolve({ payload: 'data' }))

        const { result } = renderHook(() => useParallelRequests(() => request()))

        expect(result.current.loading).toBeTruthy()
        expect(result.current.data).toBe(undefined)
        expect(result.current.error).toBe(undefined)

        // eslint-disable-next-line @typescript-eslint/require-await
        await act(async () => {
            jest.runAllTimers()
        })

        expect(result.current.loading).toBe(false)
        expect(result.current.data).toStrictEqual({ payload: 'data' })
        expect(result.current.error).toBe(undefined)
    })

    it('should execute two request one by one with queueing', async () => {
        const request1 = sinon.spy<() => Promise<{ payload: string }>>(() => Promise.resolve({ payload: 'data1' }))

        const request2 = sinon.spy<() => Promise<{ payload: string }>>(() => Promise.resolve({ payload: 'data2' }))

        const { result: result1 } = renderHook(() => useParallelRequests(() => request1()))

        const { result: result2 } = renderHook(() => useParallelRequests(() => request2()))

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
            jest.runAllTimers()
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
            jest.runAllTimers()
        })

        expect(result2.current).toStrictEqual({
            data: { payload: 'data2' },
            error: undefined,
            loading: false,
        })
    })
})
