import { useContext } from 'react'

import { Attributes, trace, Span } from '@opentelemetry/api'
import { noop } from 'lodash'

import { IS_OPEN_TELEMETRY_TRACING_ENABLED, noopSpan } from '../constants'

import { TraceContext } from './constants'

/**
 * A wrapper around `span.setAttributes()` that prefixes attribute names with `render.` string.
 * This namespacing is valuable for data exploration with tools like Honeycomb.
 */
const setRenderAttributes = (span: Span | undefined, attributes: Attributes): void => {
    if (!span) {
        return
    }

    const prefixedAttributes = Object.fromEntries(
        Object.entries(attributes).map(([key, value]) => [`render.${key}`, value])
    )
    span.setAttributes(prefixedAttributes)
}

interface UseCurrentSpanResult {
    span?: Span
    setSpanRenderAttributes: (attributes: Attributes) => void
}

/**
 * Get current OpenTelemetry tracing span from the `TraceSpanProvider` higher in the React tree.
 * Returns `noopSpan` if OpenTelemetry tracing is disabled.
 */
let useCurrentSpan = (): UseCurrentSpanResult => {
    const span = trace.getSpan(useContext(TraceContext).context)

    return {
        span,
        setSpanRenderAttributes: setRenderAttributes.bind(this, span),
    }
}

if (!IS_OPEN_TELEMETRY_TRACING_ENABLED) {
    useCurrentSpan = () => ({ span: noopSpan, setSpanRenderAttributes: noop })
}

export { useCurrentSpan }
