import { renderHook, act } from '@testing-library/react'
import type { useHistory } from 'react-router'
import sinon from 'sinon'

import { useURLSyncedState } from './useUrlSyncedState'

const replaceSpy = sinon.spy()
function useMockHistory() {
    return {
        replace: replaceSpy,
    }
}

describe('useURLSyncedState', () => {
    beforeEach(() => {
        replaceSpy.resetHistory()
    })

    it('should sync state with URL search parameters', () => {
        const searchParameters = new URLSearchParams()
        searchParameters.set('foo', 'foo')
        const { result } = renderHook(() =>
            useURLSyncedState({ bar: 'bar' }, searchParameters, useMockHistory as unknown as typeof useHistory)
        )
        const [data, setData] = result.current

        // initial state
        expect(data).toEqual({ foo: 'foo', bar: 'bar' })
        sinon.assert.calledWithExactly(replaceSpy, { search: 'bar=bar&foo=foo' })

        // on local state change
        act(() => setData({ bar: undefined }))
        const [data2] = result.current
        expect(data2).toEqual({ foo: 'foo' })
        sinon.assert.calledWithExactly(replaceSpy, { search: 'foo=foo' })
    })
})
