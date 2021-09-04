// causes false positive on act()
/* eslint-disable @typescript-eslint/no-floating-promises */

import { renderHook, act } from '@testing-library/react-hooks'

import { ThemePreference, useTheme } from './theme'

// Don't test reacting to system-wide theme changes, for simplicity. This means that
// observeSystemIsLightTheme's initial value will be used, but it will not monitor for subsequent
// changes.
jest.mock('@sourcegraph/shared/src/util/useObservable', () => ({
    useObservable: () => undefined,
}))

const mockWindow = (systemTheme: 'light' | 'dark'): Pick<Window, 'matchMedia'> => ({
    matchMedia: query => {
        if (query === '(prefers-color-scheme: dark)') {
            // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
            return { matches: systemTheme === 'dark' } as MediaQueryList
        }
        throw new Error('unexpected matchMedia query')
    },
})

const mockDocumentElement = (): Pick<HTMLElement, 'classList'> => document.createElement('html')

const mockLocalStorage = (): Pick<Storage, 'getItem' | 'setItem'> => {
    const data = new Map<string, string>()
    return {
        getItem: key => data.get(key) ?? null,
        setItem: (key, value) => {
            data.set(key, String(value))
        },
    }
}

const createUseThemeMocks = (
    systemTheme: 'light' | 'dark',
    storedThemePreference: ThemePreference | null
): Required<Parameters<typeof useTheme>> => {
    const window = mockWindow(systemTheme)
    const documentElement = mockDocumentElement()
    const storage = mockLocalStorage()
    if (storedThemePreference !== null) {
        storage.setItem('light-theme', storedThemePreference)
    }
    return [window, documentElement, storage]
}

describe('useTheme()', () => {
    describe('defaults to system', () => {
        it('light', () => {
            const [window, documentElement, localStorage] = createUseThemeMocks('light', null)
            const { result } = renderHook(() => useTheme(window, documentElement, localStorage))
            expect(result.current.isLightTheme).toBe(true)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(documentElement.classList).toContain('theme-light')
            expect(documentElement.classList).not.toContain('theme-dark')
            expect(localStorage.getItem('light-theme')).toBe(null)
        })

        it('dark', () => {
            const [window, documentElement, localStorage] = createUseThemeMocks('dark', null)
            const { result } = renderHook(() => useTheme(window, documentElement, localStorage))
            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(documentElement.classList).toContain('theme-dark')
            expect(documentElement.classList).not.toContain('theme-light')
            expect(localStorage.getItem('light-theme')).toBe(null)
        })
    })

    describe('respects theme preference', () => {
        it('light', () => {
            const [window, documentElement, localStorage] = createUseThemeMocks('dark', ThemePreference.Light)
            const { result } = renderHook(() => useTheme(window, documentElement, localStorage))
            expect(result.current.isLightTheme).toBe(true)
            expect(result.current.themePreference).toBe(ThemePreference.Light)
            expect(documentElement.classList).toContain('theme-light')
            expect(documentElement.classList).not.toContain('theme-dark')
            expect(localStorage.getItem('light-theme')).toBe(ThemePreference.Light)
        })

        it('dark', () => {
            const [window, documentElement, localStorage] = createUseThemeMocks('light', ThemePreference.Dark)
            const { result } = renderHook(() => useTheme(window, documentElement, localStorage))
            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.Dark)
            expect(documentElement.classList).toContain('theme-dark')
            expect(documentElement.classList).not.toContain('theme-light')
            expect(localStorage.getItem('light-theme')).toBe(ThemePreference.Dark)
        })

        it('system', () => {
            const [window, documentElement, localStorage] = createUseThemeMocks('dark', ThemePreference.System)
            const { result } = renderHook(() => useTheme(window, documentElement, localStorage))
            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(documentElement.classList).toContain('theme-dark')
            expect(documentElement.classList).not.toContain('theme-light')
            expect(localStorage.getItem('light-theme')).toBe(ThemePreference.System)
        })
    })

    it('changes theme preference', () => {
        const [window, documentElement, localStorage] = createUseThemeMocks('light', null)
        const { result } = renderHook(() => useTheme(window, documentElement, localStorage))
        expect(result.current.isLightTheme).toBe(true)
        expect(result.current.themePreference).toBe(ThemePreference.System)

        // Change to dark theme.
        act(() => {
            result.current.onThemePreferenceChange(ThemePreference.Dark)
        })
        expect(result.current.isLightTheme).toBe(false)
        expect(result.current.themePreference).toBe(ThemePreference.Dark)
        expect(documentElement.classList).toContain('theme-dark')
        expect(documentElement.classList).not.toContain('theme-light')
        expect(localStorage.getItem('light-theme')).toBe(ThemePreference.Dark)

        // Change to system.
        act(() => {
            result.current.onThemePreferenceChange(ThemePreference.System)
        })
        expect(result.current.isLightTheme).toBe(true)
        expect(result.current.themePreference).toBe(ThemePreference.System)
        expect(documentElement.classList).toContain('theme-light')
        expect(documentElement.classList).not.toContain('theme-dark')
        expect(localStorage.getItem('light-theme')).toBe(ThemePreference.System)
    })
})
