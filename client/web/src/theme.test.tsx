// causes false positive on act()
/* eslint-disable @typescript-eslint/no-floating-promises */
import { renderHook, act } from '@testing-library/react'

import { ThemePreference, useThemeProps } from './theme'

// Don't test reacting to system-wide theme changes, for simplicity. This means that
// observeSystemIsLightTheme's initial value will be used, but it will not monitor for subsequent
// changes.
jest.mock('@sourcegraph/wildcard', () => {
    const actual = jest.requireActual('@sourcegraph/wildcard')

    return {
        ...actual,
        useObservable: () => undefined,
    }
})

const mockSystemTheme = (systemTheme: 'light' | 'dark') => {
    window.matchMedia = query => {
        if (query === '(prefers-color-scheme: dark)') {
            // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
            return { matches: systemTheme === 'dark' } as MediaQueryList
        }
        throw new Error('unexpected matchMedia query')
    }
}

describe('useTheme()', () => {
    describe('defaults to system', () => {
        it('light', () => {
            mockSystemTheme('light')

            const { result } = renderHook(() => useThemeProps())

            expect(result.current.isLightTheme).toBe(true)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(document.documentElement.classList).toContain('theme-light')
            expect(document.documentElement.classList).not.toContain('theme-dark')
            // Local storage item not set by default, will get set when the preference is changed
            expect(localStorage).toHaveLength(0)
        })

        it('dark', () => {
            mockSystemTheme('dark')

            const { result } = renderHook(() => useThemeProps())

            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(document.documentElement.classList).toContain('theme-dark')
            expect(document.documentElement.classList).not.toContain('theme-light')
            // Local storage item not set by default, will get set when the preference is changed
            expect(localStorage).toHaveLength(0)
        })
    })

    describe('respects theme preference', () => {
        it('light', () => {
            mockSystemTheme('dark')
            const { result } = renderHook(() => useThemeProps())
            act(() => result.current.onThemePreferenceChange(ThemePreference.Light))

            expect(result.current.isLightTheme).toBe(true)
            expect(result.current.themePreference).toBe(ThemePreference.Light)
            expect(document.documentElement.classList).toContain('theme-light')
            expect(document.documentElement.classList).not.toContain('theme-dark')
            expect(JSON.parse(localStorage.getItem('temporarySettings') ?? '')).toEqual({
                'user.themePreference': 'light',
            })
        })

        it('dark', () => {
            mockSystemTheme('light')
            const { result } = renderHook(() => useThemeProps())
            act(() => result.current.onThemePreferenceChange(ThemePreference.Dark))

            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.Dark)
            expect(document.documentElement.classList).toContain('theme-dark')
            expect(document.documentElement.classList).not.toContain('theme-light')
            expect(JSON.parse(localStorage.getItem('temporarySettings') ?? '')).toEqual({
                'user.themePreference': 'dark',
            })
        })

        it('system', () => {
            mockSystemTheme('dark')
            const { result } = renderHook(() => useThemeProps())
            act(() => result.current.onThemePreferenceChange(ThemePreference.System))

            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(document.documentElement.classList).toContain('theme-dark')
            expect(document.documentElement.classList).not.toContain('theme-light')
            expect(JSON.parse(localStorage.getItem('temporarySettings') ?? '')).toEqual({
                'user.themePreference': 'system',
            })
        })
    })

    it('changes theme preference', () => {
        mockSystemTheme('light')
        const { result } = renderHook(() => useThemeProps())
        expect(result.current.isLightTheme).toBe(true)
        expect(result.current.themePreference).toBe(ThemePreference.System)

        // Change to dark theme.
        act(() => {
            result.current.onThemePreferenceChange(ThemePreference.Dark)
        })
        expect(result.current.isLightTheme).toBe(false)
        expect(result.current.themePreference).toBe(ThemePreference.Dark)
        expect(document.documentElement.classList).toContain('theme-dark')
        expect(document.documentElement.classList).not.toContain('theme-light')
        expect(JSON.parse(localStorage.getItem('temporarySettings') ?? '')).toEqual({ 'user.themePreference': 'dark' })

        // Change to system.
        act(() => {
            result.current.onThemePreferenceChange(ThemePreference.System)
        })
        expect(result.current.isLightTheme).toBe(true)
        expect(result.current.themePreference).toBe(ThemePreference.System)
        expect(document.documentElement.classList).toContain('theme-light')
        expect(document.documentElement.classList).not.toContain('theme-dark')
        expect(JSON.parse(localStorage.getItem('temporarySettings') ?? '')).toEqual({
            'user.themePreference': 'system',
        })
    })
})
