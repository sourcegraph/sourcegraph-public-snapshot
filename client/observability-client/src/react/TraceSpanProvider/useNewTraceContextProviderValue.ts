import { useMemo, useContext } from 'react'

import { trace, Span, ROOT_CONTEXT, Context, Attributes } from '@opentelemetry/api'

import { sharedSpanStore } from '../../sdk'
import { TraceContext, reactManualTracer } from '../constants'

import type { TraceSpanProviderProps } from './TraceSpanProvider'

/**
 * Use `customContext` if provided, otherwise use `parentContext`.
 * If no `parentContext` is available, use the current navigation context.
 *
 * It allows binding of all react spans to the parent span or the recent navigation event.
 */
function getRelevantContext(parentContext: Context, customContext?: Context): Context {
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

/**
 * Creates the new OpenTelemetry tracing span on the first component render call.
 * Uses span provided by the `TraceContext` as a parent span for the new span.
 */
export function useNewTraceContextProviderValue(
    options: Omit<TraceSpanProviderProps, 'children'>
): NewTraceContextProviderValue {
    const { context: providedParentContext } = useContext(TraceContext)

    return useMemo(() => {
        const { name, attributes, options: spanOptions, context: customContext } = options
        const parentContext = getRelevantContext(providedParentContext, customContext)

        const newSpan = reactManualTracer.startSpan(name, spanOptions, parentContext)
        const newContext = trace.setSpan(parentContext, newSpan)

        if (attributes) {
            setRenderAttributes(newSpan, attributes)
        }

        return {
            newSpan,
            newContext,
            traceContextProviderValue: { context: newContext },
        }
        // We want to create a new span only on the first component render call.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])
}

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
