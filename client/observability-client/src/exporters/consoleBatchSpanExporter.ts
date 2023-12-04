import type { Attributes } from '@opentelemetry/api'
import { ExportResultCode, hrTimeToMilliseconds, type ExportResult } from '@opentelemetry/core'
import type { SpanExporter, ReadableSpan } from '@opentelemetry/sdk-trace-base'

import { logger } from '@sourcegraph/common'

interface FormattedSpan {
    name?: string
    raw?: ReadableSpan
    attrs?: Attributes
    children?: FormattedSpan[]
}

/**
 * Nests spans using `parentSpanId` fields on a span batch and logs them into the console.
 *
 * Used in the development environment for a faster feedback cycle
 * and better span explorability during experimentation with new OTel instrumentations.
 */
export class ConsoleBatchSpanExporter implements SpanExporter {
    private formatDuration(span: ReadableSpan): string {
        return hrTimeToMilliseconds(span.duration).toString() + 'ms'
    }

    private groupSpans(spans: ReadableSpan[]): FormattedSpan[] {
        const rootSpans: FormattedSpan[] = []
        const formattedSpans: Record<string, FormattedSpan> = {}

        for (const span of spans) {
            const { parentSpanId, attributes } = span
            const { spanId } = span.spanContext()

            const formattedSpan: FormattedSpan = formattedSpans[spanId] ?? {
                children: [],
            }

            formattedSpan.raw = span
            formattedSpan.attrs = attributes
            formattedSpan.name = `${span.name} - ${this.formatDuration(span)}`

            if (parentSpanId) {
                const parentSpan = formattedSpans[parentSpanId]

                if (parentSpan) {
                    parentSpan.children?.push(formattedSpan)
                } else {
                    formattedSpans[parentSpanId] = { children: [formattedSpan] }
                }
            } else {
                rootSpans.push(formattedSpan)
            }

            formattedSpans[spanId] = formattedSpan
        }

        return rootSpans
    }

    private nestSpans = (span: FormattedSpan): Record<string, FormattedSpan> => {
        const { attrs, raw, children, name = 'unknown' } = span
        const spanDetails: FormattedSpan = { attrs, raw }

        if (children?.length) {
            spanDetails.children = children.map(spanChildren => this.nestSpans(spanChildren))
        }

        return {
            [name]: spanDetails,
        }
    }

    public export(spans: ReadableSpan[], resultCallback: (result: ExportResult) => void): void {
        const formattedSpans = this.groupSpans(spans).map(this.nestSpans)

        for (const span of formattedSpans) {
            logger.debug(span)
        }

        return resultCallback({ code: ExportResultCode.SUCCESS })
    }

    public shutdown(): Promise<void> {
        return Promise.resolve()
    }
}
