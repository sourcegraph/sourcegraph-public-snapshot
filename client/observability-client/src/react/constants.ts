import { createContext } from 'react'

import { ROOT_CONTEXT, trace } from '@opentelemetry/api'

export const TraceContext = createContext({ context: ROOT_CONTEXT })
export const reactManualTracer = trace.getTracer('@sourcegraph/react-manual', '0.1')
