import { trace } from '@opentelemetry/api'

export const noopTracer = trace.getTracer('noop')
export const noopSpan = noopTracer.startSpan('noop')

/**
 * When OpenTelemetry tracing is initialized, created spans
 * are active and record information like events. We need to
 * revisit this check later when we will introduce a sampler.
 *
 * See issue for more details:
 * https://github.com/open-telemetry/opentelemetry-js-api/issues/118
 */
export const IS_OPEN_TELEMETRY_TRACING_ENABLED = noopSpan.isRecording()
