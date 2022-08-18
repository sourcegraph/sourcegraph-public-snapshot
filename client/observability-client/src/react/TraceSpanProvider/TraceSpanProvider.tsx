import { PropsWithChildren, useEffect, FunctionComponent } from 'react'

import { SpanOptions, Context } from '@opentelemetry/api'

import { IS_OPEN_TELEMETRY_TRACING_ENABLED } from '../../constants'
import { TraceContext } from '../constants'

import { useNewTraceContextProviderValue } from './useNewTraceContextProviderValue'

export type TraceSpanProviderProps = PropsWithChildren<{
    name: string
    options?: SpanOptions
    context?: Context
}>

let TraceSpanProvider: FunctionComponent<TraceSpanProviderProps> = props => {
    const { children, ...restProps } = props

    const { newSpan, traceContextProviderValue } = useNewTraceContextProviderValue(restProps)

    useEffect(() => {
        newSpan.end()
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return <TraceContext.Provider value={traceContextProviderValue}>{children}</TraceContext.Provider>
}

if (!IS_OPEN_TELEMETRY_TRACING_ENABLED) {
    TraceSpanProvider = function NoopTraceSpanProvider({ children }) {
        return <>{children}</>
    }
}

export { TraceSpanProvider }
