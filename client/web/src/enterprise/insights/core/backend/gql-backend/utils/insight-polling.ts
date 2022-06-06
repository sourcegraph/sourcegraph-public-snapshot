import { BackendInsight } from '../../..'
const ALL_REPOS_POLL_INTERVAL = 10000
const SOME_REPOS_POLL_INTERVAL = 1000

export function insightPollingInterval(insight: BackendInsight): number {
    return insight.repositories.length > 0 ? ALL_REPOS_POLL_INTERVAL : SOME_REPOS_POLL_INTERVAL
}
