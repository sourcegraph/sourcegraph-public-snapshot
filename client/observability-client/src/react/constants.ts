import { createContext, MutableRefObject } from 'react'

import { Context, ROOT_CONTEXT, trace } from '@opentelemetry/api'

// Store trace context in React ref to avoid re-rendering wrapped components on trace context change.
export type TraceContextRef = MutableRefObject<{ context: Context }>

export const TraceContext = createContext<TraceContextRef>({ current: { context: ROOT_CONTEXT } })
export const reactManualTracer = trace.getTracer('@sourcegraph/react-manual', '0.1')

export enum ReactAttributes {
    ComponentName = 'react.component.name',
    ComponentPropPrefix = 'react.prop',
}
