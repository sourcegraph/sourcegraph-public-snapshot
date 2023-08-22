import { act, renderHook } from '@testing-library/react'

import { useJsoncParser } from './useJsoncParser'

describe('useJsoncParser', () => {
    it('parses, updates & resets correctly', () => {
        // initial parsing
        const { result } = renderHook(() => useJsoncParser<{ foo?: string; bar?: string }>('{"foo": "bar"}'))
        const { update, reset } = result.current
        expect(result?.current?.json).toEqual({ foo: 'bar' })
        expect(result?.current?.rawJson).toEqual('{"foo": "bar"}')
        expect(result?.current?.error).toBeUndefined()

        // update
        act(() => {
            update({ bar: 'baz' })
        })
        expect(result.current?.json).toEqual({ foo: 'bar', bar: 'baz' })
        expect(result?.current?.rawJson).toEqual('{\n  "foo": "bar",\n  "bar": "baz"\n}')

        // reset
        act(() => {
            reset()
        })
        expect(result.current?.json).toEqual({ foo: 'bar' })
        expect(result?.current?.rawJson).toEqual('{"foo": "bar"}')
    })

    it('re-parses on originalRawJson change', () => {
        // initial parsing
        const { result, rerender } = renderHook(
            ({ jsonValue }: { jsonValue: string }) => useJsoncParser<{ foo?: string; bar?: string }>(jsonValue),
            {
                initialProps: { jsonValue: '{"foo": "bar"}' },
            }
        )
        expect(result?.current?.json).toEqual({ foo: 'bar' })
        expect(result?.current?.rawJson).toEqual('{"foo": "bar"}')
        expect(result?.current?.error).toBeUndefined()

        rerender({ jsonValue: '{"bar": "baz"}' })
        expect(result?.current?.json).toEqual({ bar: 'baz' })
        expect(result?.current?.rawJson).toEqual('{"bar": "baz"}')
        expect(result?.current?.error).toBeUndefined()
    })

    it('handles parse errors', () => {
        const { result } = renderHook(() => useJsoncParser<{ foo?: string; bar?: string }>('{"foo"/ "bar"}'))
        const { json, rawJson, error } = result.current
        expect(rawJson).toEqual('{"foo"/ "bar"}')
        expect(json).toBeUndefined()
        expect(error).toBeDefined()
    })
})
