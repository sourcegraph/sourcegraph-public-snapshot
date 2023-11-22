import { useMemo, useContext } from 'react'

import { context, trace, type Span, type Context, type Attributes, ROOT_CONTEXT } from '@opentelemetry/api'

import { sharedSpanStore, areOnTheSameTrace } from '../../sdk'
import { TraceContext, reactManualTracer, ReactAttributes } from '../constants'

import type { TraceSpanProviderProps } from './TraceSpanProvider'

/**
 * This function ensures that all React spans are connected to the parent or current navigation context.
 */
function getReactTracerContext(parentContext: Context = context.active()): Context {
    const currentNavigationContext = sharedSpanStore.getRootNavigationContext() || parentContext

    // If no `parentContext` is available, use the current navigation context.
    if (parentContext === ROOT_CONTEXT) {
        return currentNavigationContext
    }

    // If `parentContext` is linked to the old `PageView` trace, use the current navigation context.
    if (!areOnTheSameTrace(trace.getSpan(currentNavigationContext), trace.getSpan(parentContext))) {
        return currentNavigationContext
    }

    return parentContext
}

interface NewTraceContextProviderValue {
    newSpan: Span
    newContext: Context
}

/**
 * Creates the new OpenTelemetry tracing span on the first component render call.
 * Uses span provided by the `TraceContext` as a parent span for the new one.
 */
export function useNewTraceContextProviderValue(
    options: Omit<TraceSpanProviderProps, 'children'>
): NewTraceContextProviderValue {
    const { context: providedParentContext } = useContext(TraceContext).current

    return useMemo(() => {
        const { name, attributes, options: spanOptions, context: customContext } = options
        const parentContext = getReactTracerContext(customContext || providedParentContext)
        const newSpan = reactManualTracer.startSpan(name, spanOptions, parentContext)

        newSpan.setAttribute(ReactAttributes.ComponentName, name)

        const newContext = trace.setSpan(parentContext, newSpan)

        if (attributes) {
            setRenderAttributes(newSpan, attributes)
        }

        return {
            newSpan,
            newContext,
        }
        // We want to create a new span only on the first component render call.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])
}

/**
 * A wrapper around `span.setAttributes()` that prefixes attribute names with
 * `ReactAttributes.ComponentPropPrefix` string. This namespacing is valuable for
 * data exploration with tools like Honeycomb.
 */
const setRenderAttributes = (span: Span | undefined, attributes: Attributes): void => {
    if (!span) {
        return
    }

    const prefixedAttributes = Object.fromEntries(
        Object.entries(attributes).map(([key, value]) => [`${ReactAttributes.ComponentPropPrefix}.${key}`, value])
    )

    span.setAttributes(prefixedAttributes)
}
