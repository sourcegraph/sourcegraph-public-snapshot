import { createContext } from 'react'

import { Attributes, ROOT_CONTEXT, Span, trace } from '@opentelemetry/api'

export const TraceContext = createContext({ context: ROOT_CONTEXT })
export const reactManualTracer = trace.getTracer('@sourcegraph/react-manual', '0.1')

export const setRenderAttributes = (span: Span, attributes: Attributes): Span => {
    const prefixedAttributes = Object.fromEntries(
        Object.entries(attributes).map(([key, value]) => [`render.${key}`, value])
    )
    span.setAttributes(prefixedAttributes)

    return span
}
