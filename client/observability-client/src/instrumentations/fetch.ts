import { context, type Context, type Tracer, ROOT_CONTEXT } from '@opentelemetry/api'

import { sharedSpanStore } from '../sdk'

/**
 * Ensures that fetch instrumentation spans are linked for the navigation span
 * in case it's in active instead of creating independent traces.
 */
function getFetchTracerContext(parentContext: Context = context.active()): Context {
    const currentNavigationContext = sharedSpanStore.getRootNavigationContext() || parentContext
    const currentNavigationSpan = sharedSpanStore.getRootNavigationSpan()

    // If no `parentContext` is available and the current navigation span is active,
    // use the current navigation context.
    if (parentContext === ROOT_CONTEXT && !currentNavigationSpan?.ended) {
        return currentNavigationContext
    }

    return parentContext
}

/**
 * Used to patch `SourcegraphWebTracerProvider` to tweak the behavior of the `@opentelemetry/instrumentation-fetch`.
 */
export const getFetchInstrumentationStartSpan =
    (originalStartSpan: Tracer['startSpan']): Tracer['startSpan'] =>
    (name, options, context) =>
        originalStartSpan(name, options, getFetchTracerContext(context))
