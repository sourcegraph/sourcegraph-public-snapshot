import { type PropsWithChildren, useEffect, type FunctionComponent, useRef } from 'react'

import { type SpanOptions, type Context, ROOT_CONTEXT } from '@opentelemetry/api'

import { IS_OPEN_TELEMETRY_TRACING_ENABLED } from '../../constants'
import { TraceContext } from '../constants'

import { useNewTraceContextProviderValue } from './useNewTraceContextProviderValue'

export type TraceSpanProviderProps = PropsWithChildren<{
    /** OpenTelemetry span name */
    name: string
    /** OpenTelemetry span options without `attributes` */
    options?: Omit<SpanOptions, 'attributes'>
    /** OpenTelemetry span attributes. Attribute names will be automatically prefixed with 'render.' */
    attributes?: SpanOptions['attributes']
    /** OpenTelemetry context */
    context?: Context
}>

/**
 * A wrapper component is used to create the OpenTelemetry tracing span
 * on the first component render call, and it ends the span automatically
 * on the component mount event. Does nothing and returns `children` if
 * OpenTelemetry tracing is disabled.
 *
 * Question: Why use React context instead of OpenTelemetry context manager?
 *
 * The OpenTelemetry context is immutable and can only be passed down
 * with a callback. If there's no way to wrap the function execution into a
 * parent span via callback we need to implement another sharing mechanism like
 * a store. This issue is raised in the OpenTelemetry repo and currently does not
 * have a recommended solution.
 *
 * Example issues:
 * 1. https://github.com/open-telemetry/opentelemetry-js-contrib/issues/995#issuecomment-1138367723
 * 2. https://github.com/open-telemetry/opentelemetry-js-contrib/issues/732
 */
let TraceSpanProvider: FunctionComponent<TraceSpanProviderProps> = props => {
    const { children, ...restProps } = props

    // Store trace context in React ref to avoid re-rendering wrapped components on trace context change.
    const traceContextProviderValueRef = useRef({ context: ROOT_CONTEXT })

    const { newSpan, newContext } = useNewTraceContextProviderValue(restProps)
    traceContextProviderValueRef.current = { context: newContext }

    useEffect(() => {
        newSpan.end()
        // The `newSpan` is only created once on the first render call.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return <TraceContext.Provider value={traceContextProviderValueRef}>{children}</TraceContext.Provider>
}

if (!IS_OPEN_TELEMETRY_TRACING_ENABLED) {
    TraceSpanProvider = function NoopTraceSpanProvider({ children }) {
        return <>{children}</>
    }
}

export { TraceSpanProvider }
