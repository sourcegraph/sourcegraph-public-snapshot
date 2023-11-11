import type { Progress, Skipped } from '@sourcegraph/shared/src/search/stream'

export const limitHit = (progress: Progress): boolean =>
    progress.skipped.some(skipped => skipped.reason.indexOf('-limit') > 0)

const severityToNumber = (severity: Skipped['severity']): number => {
    switch (severity) {
        case 'error': {
            return 1
        }
        case 'warn': {
            return 2
        }
        case 'info': {
            return 3
        }
    }
}

const severityComparer = (a: Skipped, b: Skipped): number => {
    const aSev = severityToNumber(a.severity)
    const bSev = severityToNumber(b.severity)

    return aSev - bSev
}

export const sortBySeverity = (skipped: Skipped[]): Skipped[] => skipped.slice().sort(severityComparer)
