import { renderHook, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

import { useURLSyncedString } from './useURLSyncedString'

describe('useURLSyncedString', () => {
    it('should return default value if no URL search param is set', () => {
        const { result } = renderHook(() => useURLSyncedString('foo', 'default-value'), {
            wrapper: ({ children }) => <MemoryRouter initialEntries={['']}>{children}</MemoryRouter>,
        })
        const [value, setValue] = result.current

        // initial state
        expect(value).toEqual('default-value')

        // on local state change
        act(() => setValue('baz'))
        expect(result.current[0]).toEqual('baz')
    })

    it('should sync state with URL search parameters', () => {
        const { result } = renderHook(() => useURLSyncedString('foo', 'default-value'), {
            wrapper: ({ children }) => <MemoryRouter initialEntries={['?foo=bar']}>{children}</MemoryRouter>,
        })
        const [value, setValue] = result.current

        // initial state
        expect(value).toEqual('bar')

        // on local state change
        act(() => setValue('baz'))
        expect(result.current[0]).toEqual('baz')
    })
})
