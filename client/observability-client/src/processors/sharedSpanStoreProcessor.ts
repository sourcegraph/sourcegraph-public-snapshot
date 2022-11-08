import { ReadableSpan, SpanProcessor } from '@opentelemetry/sdk-trace-base'

import { sharedSpanStore, SharedSpanName } from '../sdk'

/**
 * Saves created navigation spans to the `sharedSpanStore` for other spans
 * to used them later as parents.
 *
 * Filters spans by `span.name` using the `SharedSpanName` enum to find spans to save.
 */
export class SharedSpanStoreProcessor implements SpanProcessor {
    public onStart(span: ReadableSpan): void {
        const { name: spanName } = span

        if (Object.values(SharedSpanName).some(name => name === spanName)) {
            sharedSpanStore.set(spanName as SharedSpanName, span)
        }
    }

    // eslint-disable-next-line @typescript-eslint/no-empty-function
    public onEnd(): void {}

    public forceFlush(): Promise<void> {
        return Promise.resolve()
    }

    public shutdown(): Promise<void> {
        return Promise.resolve()
    }
}
