import type { ReadableSpan, Span } from '@opentelemetry/sdk-trace-base'

/**
 * Read/write span: A function receiving this as argument must have
 * access to both the full span API as defined in the API-level definition
 * for span's interface and additionally must be able to retrieve all information
 * that was added to the span (as with readable span).
 *
 * https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#additional-span-interfaces
 *
 * This type can be used in `SpanProcess.onStart()` method until it's fixed in the upstream.
 * https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#onstart
 *
 * See related issue:
 * https://github.com/open-telemetry/opentelemetry-js/issues/1620
 */
export type ReadWriteSpan = Span & ReadableSpan
