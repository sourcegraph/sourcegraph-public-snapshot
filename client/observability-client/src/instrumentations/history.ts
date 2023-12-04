import type { AttributeValue } from '@opentelemetry/api'

import { ClientAttributes } from '../processors/clientAttributesSpanProcessor'
import { SharedSpanName, InstrumentationBaseWeb, sharedSpanStore } from '../sdk'

const PATCHED_HISTORY_METHODS = ['replaceState', 'pushState', 'back', 'forward', 'go'] as const
type PatchedKeys = typeof PATCHED_HISTORY_METHODS[number]

interface LocationInfo {
    pathname?: AttributeValue
    search?: AttributeValue
}

interface HistoryInstrumentationOptions {
    shouldCreatePageViewOnLocationChange: (prevLocationInfo: LocationInfo) => boolean
}

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

    private shouldCreatePageViewOnLocationChange: HistoryInstrumentationOptions['shouldCreatePageViewOnLocationChange']

    constructor(options: HistoryInstrumentationOptions) {
        super(HistoryInstrumentation.instrumentationName, HistoryInstrumentation.version)

        this.shouldCreatePageViewOnLocationChange = options.shouldCreatePageViewOnLocationChange
    }

    private patchHistoryMethod = (original: History[PatchedKeys]): History[PatchedKeys] => {
        // eslint-disable-next-line unicorn/no-this-assignment, @typescript-eslint/no-this-alias
        const instrumentation = this

        return function historyMethod(this: History, ...args: unknown[]) {
            /**
             * TODO: figure out why `original as any` is required.
             * Without it the monorepo wide Typescript build fails even though
             * the `observability-client` Typescript builds are successful.
             */
            const result = (original as any).apply(this, args as any)
            instrumentation.createPageViewSpan()

            return result
        }
    }

    private createPageViewSpan = (): void => {
        const previousNavigationSpan = sharedSpanStore.getRootNavigationSpan()
        const prevLocationInfo = {
            pathname: previousNavigationSpan?.attributes[ClientAttributes.LocationPathname],
            search: previousNavigationSpan?.attributes[ClientAttributes.LocationSearch],
        }

        if (this.shouldCreatePageViewOnLocationChange(prevLocationInfo)) {
            this.createFinishedSpan({
                name: SharedSpanName.PageView,
                // Link new navigation span to the previous navigation span.
                links: previousNavigationSpan ? [{ context: previousNavigationSpan.spanContext() }] : [],
            })
        }
    }

    private patchHistoryApi(): void {
        // The spread operator in the array is required to please TS.
        this._massWrap([history], [...PATCHED_HISTORY_METHODS], this.patchHistoryMethod)
    }

    public enable(): void {
        this.patchHistoryApi()
        window.addEventListener('popstate', this.createPageViewSpan)
    }

    public disable(): void {
        // The spread operator in the array is required to please TS.
        this._massUnwrap([history], [...PATCHED_HISTORY_METHODS])
        window.removeEventListener('popstate', this.createPageViewSpan)
    }
}
