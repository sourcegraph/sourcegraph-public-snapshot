import { Span } from 'opentracing'

export async function trace<T>(name: string, parent: Span, f: (span: Span) => Promise<T>): Promise<T> {
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
