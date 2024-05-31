import { type WrapperComponent, renderHook } from '@testing-library/react-hooks'
import { describe, expect, it, vi } from 'vitest'

import type { Call } from '../client'
import { CodyProApiClientContext } from '../components/CodyProApiClient'

import { useApiCaller } from './useApiClient'

describe('useApiCaller()', () => {
    const mockCaller = {
        call: vi.fn(),
    }

    const wrapper: WrapperComponent<{ children: React.ReactNode }> = ({ children }) => (
        <CodyProApiClientContext.Provider value={{ caller: mockCaller }}>{children}</CodyProApiClientContext.Provider>
    )

    it.skip('handles successful API response', async () => {
        const mockResponse = { data: { name: 'John Doe' }, response: { ok: true, status: 200 } }
        mockCaller.call.mockResolvedValueOnce(mockResponse)
        const call: Call<unknown> = { method: 'GET', urlSuffix: '/test' }
        const { result, waitForNextUpdate } = renderHook(() => useApiCaller(call), { wrapper })

        // validate the initial state
        // loading is true because it's called immediately after the hook mounts
        expect(result.current.loading).toBe(true)
        expect(result.current.error).toBeUndefined()
        expect(result.current.data).toBeUndefined()
        expect(result.current.response).toBeUndefined()

        // wait for promise to resolve
        await waitForNextUpdate()

        expect(result.current.loading).toBe(false)
        expect(result.current.error).toBeUndefined()
        expect(result.current.data).toEqual(mockResponse.data)
        expect(result.current.response).toEqual(mockResponse.response)
    })

    it('handles generic API error response', async () => {
        const mockResponse = { data: { name: 'John Doe' }, response: { ok: false, status: 500 } }
        mockCaller.call.mockResolvedValueOnce(mockResponse)
        const call: Call<unknown> = { method: 'GET', urlSuffix: '/test' }
        const { result, waitForNextUpdate } = renderHook(() => useApiCaller(call), { wrapper })

        // wait for promise to resolve
        await waitForNextUpdate()

        expect(result.current.loading).toBe(false)
        expect(result.current.error?.message).toBe(`unexpected status code: ${mockResponse.response.status}`)
        expect(result.current.data).toBeUndefined()
        expect(result.current.response).toEqual(mockResponse.response)
    })

    it('handles 401 API error response', async () => {
        const mockResponse = { data: { name: 'John Doe' }, response: { ok: false, status: 401 } }
        mockCaller.call.mockResolvedValueOnce(mockResponse)
        const call: Call<unknown> = { method: 'GET', urlSuffix: '/test' }
        const { result, waitForNextUpdate } = renderHook(() => useApiCaller(call), { wrapper })

        // wait for promise to resolve
        await waitForNextUpdate()

        expect(result.current.loading).toBe(false)
        expect(result.current.error?.message).toBe('Please log out and log back in.')
        expect(result.current.data).toBeUndefined()
        expect(result.current.response).toEqual(mockResponse.response)
    })

    it('refetches data when refetch is called', async () => {
        const mockResponse1 = { data: { name: 'John Doe' }, response: { ok: true, status: 200 } }
        const mockResponse2 = { data: { name: 'John Doe' }, response: { ok: true, status: 200 } }
        mockCaller.call.mockResolvedValueOnce(mockResponse1).mockResolvedValueOnce(mockResponse2)
        const call: Call<unknown> = { method: 'GET', urlSuffix: '/test' }
        const { result, waitForNextUpdate } = renderHook(() => useApiCaller(call), { wrapper })

        // wait for promise to resolve
        await waitForNextUpdate()

        // validate the state after the initial API call
        expect(result.current.loading).toBe(false)
        expect(result.current.error).toBeUndefined()
        expect(result.current.data).toEqual(mockResponse1.data)
        expect(result.current.response).toEqual(mockResponse1.response)

        // call refetch
        void result.current.refetch()

        // validate loading state
        expect(result.current.loading).toBe(true)

        // wait for promise to resolve
        await waitForNextUpdate()

        // validate the state after the refetch
        expect(result.current.loading).toBe(false)
        expect(result.current.error).toBeUndefined()
        expect(result.current.data).toEqual(mockResponse1.data)
        expect(result.current.response).toEqual(mockResponse1.response)
    })

    it('handles network error', async () => {
        const errorMessage = 'Random network error'
        mockCaller.call.mockRejectedValueOnce(new Error(errorMessage))
        const call: Call<unknown> = { method: 'GET', urlSuffix: '/test' }
        const { result, waitForNextUpdate } = renderHook(() => useApiCaller(call), { wrapper })

        // wait for promise to resolve
        await waitForNextUpdate()

        expect(result.current.loading).toBe(false)
        expect(result.current.error?.message).toBe(errorMessage)
        expect(result.current.data).toBeUndefined()
        expect(result.current.response).toBeUndefined()
    })
})
