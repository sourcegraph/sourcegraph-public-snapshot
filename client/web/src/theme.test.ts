// causes false positive on act()
/* eslint-disable @typescript-eslint/no-floating-promises */

import { renderHook, act } from '@testing-library/react-hooks'

import { ThemePreference } from './stores/themeState'
import { useTheme } from './theme'

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

describe('useTheme()', () => {
    describe('defaults to system', () => {
        it('light', () => {
            window.matchMedia = mockWindow('light').matchMedia

            const { result } = renderHook(() => useTheme())

            expect(result.current.isLightTheme).toBe(true)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(document.documentElement.classList).toContain('theme-light')
            expect(document.documentElement.classList).not.toContain('theme-dark')
            expect(localStorage.getItem('light-theme')).toBe('system')
        })

        it('dark', () => {
            window.matchMedia = mockWindow('dark').matchMedia

            const { result } = renderHook(() => useTheme())

            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(document.documentElement.classList).toContain('theme-dark')
            expect(document.documentElement.classList).not.toContain('theme-light')
            expect(localStorage.getItem('light-theme')).toBe('system')
        })
    })

    describe.skip('respects theme preference', () => {
        it('light', () => {
            window.matchMedia = mockWindow('dark').matchMedia
            window.localStorage.getItem = () => 'light'

            const { result } = renderHook(() => useTheme())
            expect(result.current.isLightTheme).toBe(true)
            expect(result.current.themePreference).toBe(ThemePreference.Light)
            expect(document.documentElement.classList).toContain('theme-light')
            expect(document.documentElement.classList).not.toContain('theme-dark')
            expect(localStorage.getItem('light-theme')).toBe(ThemePreference.Light)
        })

        it('dark', () => {
            const { result } = renderHook(() => useTheme())
            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.Dark)
            expect(document.documentElement.classList).toContain('theme-dark')
            expect(document.documentElement.classList).not.toContain('theme-light')
            expect(localStorage.getItem('light-theme')).toBe(ThemePreference.Dark)
        })

        it('system', () => {
            const { result } = renderHook(() => useTheme())
            expect(result.current.isLightTheme).toBe(false)
            expect(result.current.themePreference).toBe(ThemePreference.System)
            expect(document.documentElement.classList).toContain('theme-dark')
            expect(document.documentElement.classList).not.toContain('theme-light')
            expect(localStorage.getItem('light-theme')).toBe(ThemePreference.System)
        })
    })

    it('changes theme preference', () => {
        window.matchMedia = mockWindow('light').matchMedia
        const { result } = renderHook(() => useTheme())
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
        expect(localStorage.getItem('light-theme')).toBe(ThemePreference.Dark)

        // Change to system.
        act(() => {
            result.current.onThemePreferenceChange(ThemePreference.System)
        })
        expect(result.current.isLightTheme).toBe(true)
        expect(result.current.themePreference).toBe(ThemePreference.System)
        expect(document.documentElement.classList).toContain('theme-light')
        expect(document.documentElement.classList).not.toContain('theme-dark')
        expect(localStorage.getItem('light-theme')).toBe(ThemePreference.System)
    })
})
