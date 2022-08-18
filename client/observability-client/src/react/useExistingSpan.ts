import { useContext } from 'react'

import { trace, Span } from '@opentelemetry/api'

import { IS_OPEN_TELEMETRY_TRACING_ENABLED, noopSpan } from '../constants'

import { TraceContext } from './constants'

let useExistingSpan = (): Span => {
    const span = trace.getSpan(useContext(TraceContext).context)

    if (!span) {
        throw new Error(
            'Tried to access non-existent span! Make sure that `TraceSpanProvider` is defined higher in the React tree.'
        )
    }

    return span
}

if (!IS_OPEN_TELEMETRY_TRACING_ENABLED) {
    useExistingSpan = () => noopSpan
}

export { useExistingSpan }
