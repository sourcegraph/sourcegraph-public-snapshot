import { type BackendInsight, isComputeInsight } from '../../..'

const ALL_REPOS_POLL_INTERVAL = 30000
const SOME_REPOS_POLL_INTERVAL = 7000

export function insightPollingInterval(insight: BackendInsight): number {
    if (isComputeInsight(insight)) {
        return SOME_REPOS_POLL_INTERVAL
    }

    if (insight.repoQuery.trim() === 'repo:.*') {
        return ALL_REPOS_POLL_INTERVAL
    }

    return SOME_REPOS_POLL_INTERVAL
}
