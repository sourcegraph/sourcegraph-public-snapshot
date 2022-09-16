import { Span } from '@opentelemetry/api'
import { SpanProcessor } from '@opentelemetry/sdk-trace-base'

import { getBrowserName } from '@sourcegraph/common'

export enum ClientAttributes {
    LocationHref = 'window.location.href',
    LocationPathname = 'window.location.pathname',
    LocationSearch = 'window.location.search',
    AppVersion = 'app.version',
    BrowserName = 'browser.name',
}

/**
 * Adds span attributes applicable to every span created on the client.
 */
export class ClientAttributesSpanProcessor implements SpanProcessor {
    constructor(private version: string) {}

    public onStart(span: Span): void {
        span.setAttributes({
            [ClientAttributes.LocationHref]: location.href,
            [ClientAttributes.LocationPathname]: location.pathname,
            [ClientAttributes.LocationSearch]: location.search,
            [ClientAttributes.AppVersion]: this.version,
            [ClientAttributes.BrowserName]: getBrowserName(),
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
