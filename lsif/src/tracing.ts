import * as LightStep from 'lightstep-tracer'
import { Configuration } from './config'
import { Span, Tracer } from 'opentracing'
import { initTracerFromEnv } from 'jaeger-client'

/**
 * Create a distributed tracer.
 *
 * @param serviceName The name of the process.
 * @param configuration The current configuration.
 */
export function createTracer(serviceName: string, configuration: Configuration): Tracer | undefined {
    if (configuration.useJaeger) {
        const config = {
            serviceName,
            sampler: {
                type: 'const',
                param: 1,
            },
        }

        return initTracerFromEnv(config, {})
    }

    if (configuration.lightstepAccessToken !== '') {
        return new LightStep.Tracer({
            access_token: configuration.lightstepAccessToken,
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
