import * as LightStep from 'lightstep-tracer'
import { ConfigurationContext } from './config'
import { Span, Tracer } from 'opentracing'
import { initTracerFromEnv } from 'jaeger-client'

/**
 * Create a distributed tracer.
 *
 * @param serviceName The name of the process.
 * @param ctx The configuration context instance.
 */
export function createTracer(serviceName: string, ctx: ConfigurationContext): Tracer | undefined {
    if (ctx.current.useJaeger) {
        return initTracerFromEnv(
            {
                serviceName,
                sampler: { type: 'const', param: 1 },
            },
            {}
        )
    }

    if (ctx.current.lightstepAccessToken !== '') {
        return new LightStep.Tracer({
            access_token: ctx.current.lightstepAccessToken,
            component_name: serviceName,
        })
    }

    return undefined
}

/**
 * Trace an operation.
 *
 * @param name The log message to output.
 * @param parent The parent span instance.
 * @param f The operation to perform.
 */
export async function trace<T>(name: string, parent: Span, f: (span: Span) => Promise<T> | T): Promise<T> {
    const span = parent.tracer().startSpan(name, { childOf: parent })

    try {
        return await f(span)
    } catch (error) {
        span.setTag('error', true)
        span.log({
            event: 'error',
            'error.object': error,
            stack: error.stack,
            message: error.message,
        })

        throw error
    } finally {
        span.finish()
    }
}
