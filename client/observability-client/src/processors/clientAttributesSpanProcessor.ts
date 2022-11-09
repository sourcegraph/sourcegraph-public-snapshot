import { Span } from '@opentelemetry/api'
import { hrTimeToMilliseconds } from '@opentelemetry/core'
import { ReadableSpan, SpanProcessor } from '@opentelemetry/sdk-trace-base'
import { SemanticAttributes } from '@opentelemetry/semantic-conventions'

import { getBrowserName } from '@sourcegraph/common'

import { areOnTheSameTrace, isNavigationSpanName, isSharedSpanName, sharedSpanStore } from '../sdk'

export enum ClientAttributes {
    LocationHref = 'window.location.href',
    LocationPathname = 'window.location.pathname',
    LocationSearch = 'window.location.search',
    AppVersion = 'app.version',
    BrowserName = 'browser.name',
    TimeSinceNavigation = 'time.since_navigation',
    TimeSinceAppMount = 'time.since_app_mount',
}

/**
 * Adds span attributes applicable to every span created on the client.
 */
export class ClientAttributesSpanProcessor implements SpanProcessor {
    constructor(private version: string) {}

    public onStart(span: Span): void {
        const { startTime, name } = (span as unknown) as ReadableSpan
        const startTimeMs = hrTimeToMilliseconds(startTime)
        const appMountSpan = sharedSpanStore.getAppMountSpan()
        const navigationSpan = sharedSpanStore.getRootNavigationSpan()

        if (navigationSpan && !isNavigationSpanName(name) && areOnTheSameTrace(span, navigationSpan)) {
            span.setAttribute(
                ClientAttributes.TimeSinceNavigation,
                startTimeMs - hrTimeToMilliseconds(navigationSpan.startTime)
            )
        }

        if (appMountSpan && !isSharedSpanName(name) && areOnTheSameTrace(span, appMountSpan)) {
            span.setAttribute(
                ClientAttributes.TimeSinceAppMount,
                startTimeMs - hrTimeToMilliseconds(appMountSpan.startTime)
            )
        }

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
}
