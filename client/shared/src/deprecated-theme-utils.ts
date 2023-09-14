import { concat, defer, fromEvent, type Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

/**
 * Returns an Observable that emits the system color scheme using a `prefers-color-scheme` media
 * query. The Observable will emit with the initial value immediately. Callers that need the initial
 * value synchronously can use initialValue.
 *
 * @param window_ The global window object (or a mock in tests).
 *
 * @deprecated Use useTheme and useSystemTheme instead.
 */
export const observeSystemIsLightTheme = (
    window_: Pick<Window, 'matchMedia'> = window
): { observable: Observable<boolean>; initialValue: boolean } => {
    const mediaList = window_.matchMedia('(prefers-color-scheme: dark)')
    return {
        observable: concat(
            // We want every subscriber to get the _current_ match value, hence
            // we defer evaluation of until subscription.
            defer(() => of(!mediaList.matches)),
            fromEvent<MediaQueryListEvent>(mediaList, 'change').pipe(map(event => !event.matches))
        ),
        initialValue: !mediaList.matches,
    }
}
