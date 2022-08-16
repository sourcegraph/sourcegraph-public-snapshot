import { useMemo, useContext } from 'react'

import { trace, Span, ROOT_CONTEXT, Context } from '@opentelemetry/api'

import { sharedSpanStore } from '../../sdk'
import { TraceContext, reactManualTracer } from '../constants'

import type { TraceSpanProviderProps } from './TraceSpanProvider'

function getFinalContext(parentContext: Context, customContext?: Context): Context {
    if (customContext) {
        return customContext
    }

    if (parentContext === ROOT_CONTEXT) {
        return sharedSpanStore.getRootNavigationContext() || ROOT_CONTEXT
    }

    return parentContext
}

interface NewTraceContextProviderValue {
    newSpan: Span
    newContext: Context
    traceContextProviderValue: { context: Context }
}

export function useNewTraceContextProviderValue(
    options: Omit<TraceSpanProviderProps, 'children'>
): NewTraceContextProviderValue {
    const { context: providedParentContext } = useContext(TraceContext)

    return useMemo(() => {
        const { name, options: spanOptions, context: customContext } = options
        const parentContext = getFinalContext(providedParentContext, customContext)

        const newSpan = reactManualTracer.startSpan(name, spanOptions, parentContext)
        const newContext = trace.setSpan(parentContext, newSpan)

        return {
            newSpan,
            newContext,
            traceContextProviderValue: { context: newContext },
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])
}
