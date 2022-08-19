import { otperformance } from '@opentelemetry/core'

import { SharedSpanName, ActiveSpanConfig, InstrumentationBaseWeb, sharedSpanStore } from '../sdk'

const PATCHED_HISTORY_METHODS = ['replaceState', 'pushState', 'back', 'forward', 'go'] as const
type PatchedKeys = typeof PATCHED_HISTORY_METHODS[number]

/**
 * Auto instrumentation of the window `popstate` event and history API that
 * creates the `PageView` span on every URL change. These navigation spans are
 * used as parents for other spans created by the application.
 *
 * Having top-level navigation allows grouping other spans in one trace bound to
 * the page view, which helps analyze data in Honeycomb. This experimental approach
 * will be improved based on our experience with events received from the production environment.
 *
 * Implementation is based on the `opentelemetry-instrumentation-user-interaction` instrumentation:
 * https://github.com/open-telemetry/opentelemetry-js-contrib/blob/main/plugins/web/opentelemetry-instrumentation-user-interaction/src/instrumentation.ts#L377
 */
export class HistoryInstrumentation extends InstrumentationBaseWeb {
    public static instrumentationName = '@sourcegraph/instrumentation-history'
    public static version = '0.1'

    constructor() {
        super(HistoryInstrumentation.instrumentationName, HistoryInstrumentation.version)
    }

    private patchHistoryMethod = (original: History[PatchedKeys]): History[PatchedKeys] => {
        // eslint-disable-next-line unicorn/no-this-assignment, @typescript-eslint/no-this-alias
        const instrumentation = this

        return function historyMethod(this: History, ...args: unknown[]) {
            const navigationSpan = sharedSpanStore.getRootNavigationSpan()

            const spanConfig: ActiveSpanConfig = {
                name: SharedSpanName.PageView,
                // Link new navigation span to the previous navigation span.
                links: navigationSpan ? [{ context: navigationSpan.spanContext() }] : [],
            }

            return instrumentation.createActiveSpan(spanConfig, pageViewSpan => {
                const result = original.apply(this, args as any)
                pageViewSpan.end()

                return result
            })
        }
    }

    private handlePopState = (): void => {
        this.createFinishedSpan({
            name: SharedSpanName.PageView,
            startTime: otperformance.now(),
        })
    }

    private patchHistoryApi(): void {
        // The spread operator in the array is required to please TS.
        this._massWrap([history], [...PATCHED_HISTORY_METHODS], this.patchHistoryMethod)
    }

    public enable(): void {
        this.patchHistoryApi()
        window.addEventListener('popstate', this.handlePopState)
    }

    public disable(): void {
        // The spread operator in the array is required to please TS.
        this._massUnwrap([history], [...PATCHED_HISTORY_METHODS])
        window.removeEventListener('popstate', this.handlePopState)
    }
}
