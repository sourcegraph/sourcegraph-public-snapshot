import { Observable } from 'rxjs/Observable'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { ReplaySubject } from 'rxjs/ReplaySubject'
import { eventLogger } from '../tracking/eventLogger'

/**
 *  All available color themes.
 */
export type ColorTheme = 'light' | 'dark'

/**
 * Returns the active color theme.
 */
export function getColorTheme(): ColorTheme {
    return window.localStorage.getItem('light-theme') === 'true' ? 'light' : 'dark'
}

/**
 * Sets the active color theme.
 */
export function setColorTheme(theme: ColorTheme): void {
    const isLightTheme = theme === 'light'
    window.localStorage.setItem('light-theme', isLightTheme.toString())
    colorThemeUpdates.next(theme)
    eventLogger.log(theme === 'light' ? 'LightThemeClicked' : 'DarkThemeClicked')
}

const colorThemeUpdates = new ReplaySubject<ColorTheme>(1)

// Populate with initial value.
colorThemeUpdates.next(getColorTheme())

/**
 * Represents the latest state of the color theme setting.
 */
export const colorTheme: Observable<ColorTheme> = colorThemeUpdates.pipe(distinctUntilChanged())
