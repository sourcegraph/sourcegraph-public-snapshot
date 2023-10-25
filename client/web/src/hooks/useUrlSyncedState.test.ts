import { renderHook, act } from '@testing-library/react'
import type { useNavigate } from 'react-router-dom'
import sinon from 'sinon'
import { beforeEach, describe, expect, it } from 'vitest'

import { useURLSyncedState } from './useUrlSyncedState'

const navigateSpy = sinon.spy()
function useMockNavigate() {
    return navigateSpy
}

describe('useURLSyncedState', () => {
    beforeEach(() => {
        navigateSpy.resetHistory()
    })

    it('should sync state with URL search parameters', () => {
        const searchParameters = new URLSearchParams()
        searchParameters.set('foo', 'foo')
        const { result } = renderHook(() =>
            useURLSyncedState({ bar: 'bar' }, searchParameters, useMockNavigate as unknown as typeof useNavigate)
        )
        const [data, setData] = result.current

        // initial state
        expect(data).toEqual({ foo: 'foo', bar: 'bar' })
        sinon.assert.calledWithExactly(navigateSpy, { search: 'bar=bar&foo=foo' }, { replace: true })

        // on local state change
        act(() => setData({ bar: undefined }))
        const [data2] = result.current
        expect(data2).toEqual({ foo: 'foo' })
        sinon.assert.calledWithExactly(navigateSpy, { search: 'foo=foo' }, { replace: true })
    })
})
