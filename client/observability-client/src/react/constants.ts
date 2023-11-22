import { createContext, type MutableRefObject } from 'react'

import { type Context, ROOT_CONTEXT, trace } from '@opentelemetry/api'

// Store trace context in React ref to avoid re-rendering wrapped components on trace context change.
export type TraceContextRef = MutableRefObject<{ context: Context }>

export const TraceContext = createContext<TraceContextRef>({ current: { context: ROOT_CONTEXT } })

export const REACT_TRACER_NAME = '@sourcegraph/react-manual'
export const reactManualTracer = trace.getTracer(REACT_TRACER_NAME, '0.1')

export enum ReactAttributes {
    ComponentName = 'react.component.name',
    ComponentPropPrefix = 'react.prop',
}
