import React from 'react'

import { renderHook } from '@testing-library/react-hooks'
import { describe, expect, it } from 'vitest'

import { Call, Caller } from '../client'
import { CodyProApiClientContext } from '../components/CodyProApiClient'

import { useApiCaller } from './useApiClient'

// FakeCaller is a testing fake for the Caller interface, for simulating
// making API calls. Only supports one call being made at a time, otherwise
// will fail.
//
// It's hard to do async hook testing correctly. This might be helpful:
// https://react-hooks-testing-library.com/usage/advanced-hooks#async
class FakeCaller implements Caller {
    private callInFlight: boolean = false
    private resolveLastCallFn: any | undefined = undefined
    private rejectLastCallFn: any | undefined = undefined

    call<Data>(_: Call<Data>): Promise<{ data?: Data; response: Response }> {
        if (this.callInFlight) {
            throw Error('There is already a call in-flight. You must call `reset()`')
        }

        return new Promise<{ data?: Data; response: Response }>((resolve, reject) => {
            this.callInFlight = true
            this.resolveLastCallFn = resolve
            this.rejectLastCallFn = reject

            // We leave the promise in this running state,
            // requiring the testcase to call resolveLastCallWith.
        })
    }

    isCallInFlight(): boolean {
        return this.callInFlight
    }

    resolveLastCallWith<Data>(result: { data?: Data; response: Response }) {
        if (!this.resolveLastCallFn) {
            throw Error('Cannot resolve. There is no call in-flight.')
        }
        this.resolveLastCallFn(result)
        this.reset()
    }

    rejectLastCallWith(reason: any) {
        if (!this.rejectLastCallFn) {
            throw Error('Cannot reject. There is no call in-flight.')
        }
        this.rejectLastCallFn(reason)
        this.reset()
    }

    reset() {
        if (!this.callInFlight) {
            throw Error('Cannot reset. There is no call in-flight')
        }
        this.callInFlight = false
        this.resolveLastCallFn = undefined
        this.rejectLastCallFn = undefined
    }
}

describe('useApiCaller()', () => {
    const mockCaller = new FakeCaller()
    const wrapper = ({ children }: { children: React.ReactNode }) => (
        <CodyProApiClientContext.Provider value={{ caller: mockCaller }}>{children}</CodyProApiClientContext.Provider>
    )

    // responseStub is a stubbed out Response object.
    const responseStub: Response = {
        status: 200,
    } as any

    it('works', async () => {
        const call: Call<void> = {
            method: 'GET',
            urlSuffix: '/test',
        }

        // Verify the initial state is loading.
        const { result, waitForNextUpdate } = renderHook(() => useApiCaller(call), { wrapper })
        {
            let { loading, error, data } = result.current
            expect(loading).toBe(true)
            expect(data).toBeUndefined()
            expect(error).toBeUndefined()
        }

        // Resolve the promise that was returned by the API call made by
        // the useApiCaller hook.
        expect(mockCaller.isCallInFlight()).toBe(true)
        mockCaller.resolveLastCallWith({ data: 'some value', response: responseStub })
        expect(mockCaller.isCallInFlight()).toBe(false)

        // Now we need to kick the React runtime to pick up on the change.
        await waitForNextUpdate()

        // Verify the updated state has the result from the caller.
        {
            let { loading, error, data } = result.current
            expect(loading).toBe(false)
            expect(data).toBe('some value')
            expect(error).toBeUndefined()
        }
    })

    it('handles runtime errors', async () => {
        const call: Call<void> = {
            method: 'GET',
            urlSuffix: '/test',
        }

        // Verify the initial state is loading.
        const { result, waitForNextUpdate } = renderHook(() => useApiCaller(call), { wrapper })
        {
            let { loading, error, data } = result.current
            expect(loading).toBe(true)
            expect(data).toBeUndefined()
            expect(error).toBeUndefined()
        }

        mockCaller.rejectLastCallWith(Error('Random Network Error'))
        await waitForNextUpdate()

        // Verify the error field is set
        {
            let { loading, error, data } = result.current
            expect(loading).toBe(false)
            expect(data).toBeUndefined()
            expect(error).toBeTruthy()
            expect(error?.message).toBe('Random Network Error')
        }
    })
})
