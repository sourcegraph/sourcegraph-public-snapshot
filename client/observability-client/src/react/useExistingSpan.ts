import { useContext } from 'react'

import { trace, Span } from '@opentelemetry/api'

import { TraceContext } from './constants'

export const useExistingSpan = (): Span => {
    const span = trace.getSpan(useContext(TraceContext).context)

    if (!span) {
        throw new Error(
            'Tried to access non-existent span! Make sure that `TraceSpanProvider` is defined higher in the React tree.'
        )
    }

    return span
}
