import isAbsoluteUrl from 'is-absolute-url'

export const TRACING_URL_SUFFIX = 'v1/traces'

export const getTracingURL = (endpoint: string, externalURL: string): string => {
    const url = isAbsoluteUrl(endpoint) ? endpoint : new URL(endpoint, externalURL).toString()

    // Ensure trailing slash to correctly add `TRACING_URL_SUFFIX` on the next step.
    const urlWithEndingSlash = url.replace(/\/?$/, '/')

    // As per spec non-signal-specific configuration should have signal-specific paths appended.
    // https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#endpoint-urls-for-otlphttp
    return new URL(TRACING_URL_SUFFIX, urlWithEndingSlash).toString()
}
