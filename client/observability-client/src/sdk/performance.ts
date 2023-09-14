import type { Span } from '@opentelemetry/api'
import { otperformance } from '@opentelemetry/core'
import { hasKey, type PerformanceEntries, PerformanceTimingNames } from '@opentelemetry/sdk-trace-web'
import { camelCase } from 'lodash'

/**
 * Picks `navigation` performance entries matching keys specified in `PerformanceTimingNames`.
 *
 * See https://developer.mozilla.org/en-US/docs/Web/API/PerformanceNavigationTiming
 */
export function performanceNavigationTimingToEntries(): PerformanceEntries {
    const [timing] = otperformance.getEntriesByType('navigation') as PerformanceNavigationTiming[]

    return Object.values(PerformanceTimingNames).reduce<PerformanceEntries>((result, key) => {
        if (timing && hasKey(timing, key)) {
            const value = timing[key]

            if (typeof value === 'number') {
                result[key] = value
            }
        }

        return result
    }, {})
}

/**
 * If `paint` performance entries are available adds `first-paint`
 * and `first-contentful-paint` events if to the span.
 *
 * See https://developer.mozilla.org/en-US/docs/Web/API/PerformancePaintTiming
 */
export const addSpanPerformancePaintEvents = (span: Span): void => {
    for (const { name, startTime } of otperformance.getEntriesByType('paint')) {
        span.addEvent(camelCase(name), startTime)
    }
}
