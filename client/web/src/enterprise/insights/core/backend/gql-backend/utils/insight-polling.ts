import { BackendInsight } from '../../..'

const ALL_REPOS_POLL_INTERVAL = 30000
const SOME_REPOS_POLL_INTERVAL = 2000

export function insightPollingInterval(insight: BackendInsight): number {
    return insight.repositories.length > 0 ? SOME_REPOS_POLL_INTERVAL : ALL_REPOS_POLL_INTERVAL
}
