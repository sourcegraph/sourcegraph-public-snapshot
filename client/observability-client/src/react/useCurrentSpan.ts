import { useContext } from 'react'

import { trace, type Span } from '@opentelemetry/api'

import { IS_OPEN_TELEMETRY_TRACING_ENABLED, noopSpan } from '../constants'

import { TraceContext } from './constants'

interface UseCurrentSpanResult {
    span?: Span
}

/**
 * Get current OpenTelemetry tracing span from the `TraceSpanProvider` higher in the React tree.
 * Returns `noopSpan` if OpenTelemetry tracing is disabled.
 */
let useCurrentSpan = (): UseCurrentSpanResult => {
    const span = trace.getSpan(useContext(TraceContext).current.context)

    return {
        span,
    }
}

if (!IS_OPEN_TELEMETRY_TRACING_ENABLED) {
    useCurrentSpan = () => ({ span: noopSpan })
}

export { useCurrentSpan }
