import { act, renderHook } from '@testing-library/react'
import Cookies from 'js-cookie'
import { afterEach, describe, expect, it } from 'vitest'

import { useCookieStorage } from './useCookieStorage'

describe('useCookieStorage', () => {
    afterEach(() => {
        Cookies.remove('test')
    })

    describe('typeof value === "string"', () => {
        it('should get initial value', () => {
            const { result } = renderHook(() => useCookieStorage('test', 'initial'))
            expect(result.current[0]).toEqual('initial')
            expect(Cookies.get('test')).toEqual(JSON.stringify('initial'))
        })

        it('should set and get value', () => {
            const { result } = renderHook(() => useCookieStorage('test'))
            act(() => {
                result.current[1]('value')
            })
            expect(result.current[0]).toEqual('value')
            expect(Cookies.get('test')).toEqual(JSON.stringify('value'))
        })

        it('should remove value', () => {
            const { result } = renderHook(() => useCookieStorage('test', 'value'))
            act(() => {
                result.current[1](undefined)
            })
            expect(result.current[0]).toEqual(undefined)
            expect(Cookies.get('test')).toEqual(undefined)
        })
    })
    describe('typeof value === "object"', () => {
        it('should get initial value', () => {
            const { result } = renderHook(() => useCookieStorage('test', { initial: true }))
            expect(result.current[0]).toEqual({ initial: true })
            expect(Cookies.get('test')).toEqual(JSON.stringify({ initial: true }))
        })

        it('should set and get value', () => {
            const { result } = renderHook(() => useCookieStorage('test'))
            act(() => {
                result.current[1]({ value: true })
            })
            expect(result.current[0]).toEqual({ value: true })
            expect(Cookies.get('test')).toEqual(JSON.stringify({ value: true }))
        })

        it('should remove value', () => {
            const { result } = renderHook(() => useCookieStorage('test', { value: true }))
            act(() => {
                result.current[1](undefined)
            })
            expect(result.current[0]).toEqual(undefined)
            expect(Cookies.get('test')).toEqual(undefined)
        })
    })

    describe.each([true, false])('typeof value === "boolean" && value === %s', value => {
        it('should get initial value', () => {
            const { result } = renderHook(() => useCookieStorage('test', value))
            expect(result.current[0]).toEqual(value)
            expect(Cookies.get('test')).toEqual(JSON.stringify(value))
        })

        it('should set and get value', () => {
            const { result } = renderHook(() => useCookieStorage('test'))
            act(() => {
                result.current[1](value)
            })
            expect(result.current[0]).toEqual(value)
            expect(Cookies.get('test')).toEqual(JSON.stringify(value))
        })

        it('should remove value', () => {
            const { result } = renderHook(() => useCookieStorage('test', value))
            act(() => {
                result.current[1](undefined)
            })
            expect(result.current[0]).toEqual(undefined)
            expect(Cookies.get('test')).toEqual(undefined)
        })
    })
})
