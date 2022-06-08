import { BADGE_VARIANTS } from '@sourcegraph/wildcard'

export function scoreToClassSuffix(score: number): typeof BADGE_VARIANTS[number] {
    return score > 8 ? 'success' : score > 6 ? 'info' : 'danger'
}
