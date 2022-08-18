import { useContext } from 'react'

import { trace, Span } from '@opentelemetry/api'

import { IS_OPEN_TELEMETRY_TRACING_ENABLED, noopSpan } from '../constants'

import { TraceContext } from './constants'

/**
 * Get existing OpenTelemetry tracing span from the `TraceSpanProvider` higher in the React tree.
 * Returns `noopSpan` if OpenTelemetry tracing is disabled.
 */
let useExistingSpan = (): Span | undefined => trace.getSpan(useContext(TraceContext).context)

if (!IS_OPEN_TELEMETRY_TRACING_ENABLED) {
    useExistingSpan = () => noopSpan
}

export { useExistingSpan }
