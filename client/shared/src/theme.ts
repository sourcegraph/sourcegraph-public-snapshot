import { concat, fromEvent, Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

/**
 * Props that can be extended by any component's Props which needs to react to theme change.
 */
export interface ThemeProps {
    /**
     * `true` if the current theme to be shown is the light theme,
     * `false` if it is the dark theme.
     */
    isLightTheme: boolean
}

/**
 * Returns an Observable that emits the system color scheme using a `prefers-color-scheme` media query.
 * The Observable will emit with the initial value immediately.
 */
export const observeSystemIsLightTheme = (): Observable<boolean> => {
    const mediaList = window.matchMedia('(prefers-color-scheme: dark)')
    return concat(
        of(!mediaList.matches),
        fromEvent<MediaQueryListEvent>(mediaList, 'change').pipe(map(event => !event.matches))
    )
}
