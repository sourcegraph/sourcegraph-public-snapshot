import { createContext } from 'react'

import { Attributes, ROOT_CONTEXT, Span, trace } from '@opentelemetry/api'

export const TraceContext = createContext({ context: ROOT_CONTEXT })
export const reactManualTracer = trace.getTracer('@sourcegraph/react-manual', '0.1')

/**
 * A wrapper around `span.setAttributes()` that prefixes attribute names with `render.` string.
 * This namespacing is valuable for data exploration with tools like Honeycomb.
 */
export const setRenderAttributes = (span: Span | undefined, attributes: Attributes): void => {
    if (!span) {
        return
    }

    const prefixedAttributes = Object.fromEntries(
        Object.entries(attributes).map(([key, value]) => [`render.${key}`, value])
    )
    span.setAttributes(prefixedAttributes)
}
