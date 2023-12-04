import { hrTimeToMilliseconds } from '@opentelemetry/core'
import type { SpanProcessor } from '@opentelemetry/sdk-trace-base'
import { SemanticAttributes } from '@opentelemetry/semantic-conventions'

import { getBrowserName } from '@sourcegraph/common'

import {
    areOnTheSameTrace,
    isNavigationSpanName,
    isSharedSpanName,
    SharedSpanName,
    sharedSpanStore,
    type ReadWriteSpan,
} from '../sdk'

export enum ClientAttributes {
    /**
     * window.location information.
     */
    LocationHref = 'window.location.href',
    LocationPathname = 'window.location.pathname',
    LocationSearch = 'window.location.search',
    PreviousLocationHref = 'window.prev_location.href',

    /**
     * Application specific information.
     */
    AppVersion = 'app.version',

    /**
     * Browser information.
     */
    BrowserName = 'browser.name',

    /**
     * Precomputed attributes used to build Honeycomb dashboards.
     */
    TimeSinceWindowLoad = 'time.since_window_load',
    TimeSincePageView = 'time.since_page_view',
    TimeSinceAppMount = 'time.since_app_mount',
}

/**
 * Adds span attributes applicable to every span created on the client.
 */
export class ClientAttributesSpanProcessor implements SpanProcessor {
    constructor(private version: string) {}

    public onStart(span: ReadWriteSpan): void {
        this.setTimeSinceAttributes(span)
        this.setPreviousLocationAttributes(span)

        span.setAttributes({
            [ClientAttributes.LocationHref]: location.href,
            [ClientAttributes.LocationPathname]: location.pathname,
            [ClientAttributes.LocationSearch]: location.search,
            [ClientAttributes.AppVersion]: this.version,
            [ClientAttributes.BrowserName]: getBrowserName(),
            [SemanticAttributes.HTTP_USER_AGENT]: navigator.userAgent,
        })
    }

    // eslint-disable-next-line @typescript-eslint/no-empty-function
    public onEnd(): void {}

    public forceFlush(): Promise<void> {
        return Promise.resolve()
    }

    public shutdown(): Promise<void> {
        return Promise.resolve()
    }

    /**
     * Calculate the time elapsed since the recent navigation and attach it span attributes.
     * Allows querying this data in Honeycomb because it's impossible to calculate it there on demand.
     */
    private setTimeSinceAttributes(span: ReadWriteSpan): void {
        const { startTime, name } = span
        const startTimeMs = hrTimeToMilliseconds(startTime)
        const appMountSpan = sharedSpanStore.getAppMountSpan()
        const navigationSpan = sharedSpanStore.getRootNavigationSpan()

        /**
         * Add time since recent `PageView` or `WindowLoad` span start.
         */
        if (navigationSpan && !isNavigationSpanName(name) && areOnTheSameTrace(span, navigationSpan)) {
            const timeSinceNavigationSpanName =
                navigationSpan.name === SharedSpanName.PageView
                    ? ClientAttributes.TimeSincePageView
                    : ClientAttributes.TimeSinceWindowLoad

            span.setAttribute(timeSinceNavigationSpanName, startTimeMs - hrTimeToMilliseconds(navigationSpan.startTime))
        }

        /**
         * Add time since recent `AppMount` span start.
         */
        if (appMountSpan && !isSharedSpanName(name) && areOnTheSameTrace(span, appMountSpan)) {
            span.setAttribute(
                ClientAttributes.TimeSinceAppMount,
                startTimeMs - hrTimeToMilliseconds(appMountSpan.startTime)
            )
        }
    }

    /**
     * Attach the previous `location.href` to every span to make this data available for
     * Honeycomb queries. Helpful in querying spans started upon leaving a specific part
     * of the web application.
     */
    private setPreviousLocationAttributes(span: ReadWriteSpan): void {
        const { name } = span
        const navigationSpan = sharedSpanStore.getRootNavigationSpan()

        if (isNavigationSpanName(name)) {
            const prevLocationHref = navigationSpan?.attributes[ClientAttributes.LocationHref]

            // For the navigation span set the previous location from the previous navigation span.
            if (prevLocationHref) {
                span.setAttribute(ClientAttributes.PreviousLocationHref, prevLocationHref)
            }
        } else {
            const prevLocationHref = navigationSpan?.attributes[ClientAttributes.PreviousLocationHref]

            // For non navigation spans use the `PreviousLocationHref` save in the current navigation span.
            if (prevLocationHref) {
                span.setAttribute(ClientAttributes.PreviousLocationHref, prevLocationHref)
            }
        }
    }
}
