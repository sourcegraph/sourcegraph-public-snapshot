import { renderHook, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import { useURLSyncedString } from './useUrlSyncedString'

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

    // it should not override other useURLSyncedString states
    it('should not override other useURLSyncedString states', () => {
        const { result } = renderHook(
            () => ({
                foo: useURLSyncedString('foo', 'default-foo'),
                bar: useURLSyncedString('bar', 'default-bar'),
            }),
            {
                wrapper: ({ children }) => (
                    <MemoryRouter initialEntries={['?foo=foo1&bar=bar1']}>{children}</MemoryRouter>
                ),
            }
        )

        // initial state
        expect(result.current.foo[0]).toEqual('foo1')
        expect(result.current.bar[0]).toEqual('bar1')

        // on local state change
        act(() => result.current.foo[1]('foo2'))
        expect(result.current.foo[0]).toEqual('foo2')
        expect(result.current.bar[0]).toEqual('bar1')
    })
})
