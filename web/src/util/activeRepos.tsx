import { fetchActiveRepos } from 'sourcegraph/backend'
import { ActiveRepoResults } from 'sourcegraph/util/types'

const localStorageKey = 'activeRepos'

interface LocalStorage extends GQL.IActiveRepoResults {
    timestamp: number
}

/**
 * get returns the ActiveRepoResults and properly fetches / caches them in
 * local storage.
 */
export function get(): Promise<ActiveRepoResults> {
    // Uncomment to debug the non-cached path more easily:
    // window.localStorage.setItem(localStorageKey, "");

    let activeRepos: LocalStorage
    const data = window.localStorage.getItem(localStorageKey)
    if (data) {
        activeRepos = JSON.parse(data)
        const halfHour = 30 * 60 * 1000 // 30m * 60s * 1000ms == 30m in milliseconds
        if (activeRepos.timestamp && (Date.now() - activeRepos.timestamp) < halfHour) {
            // data exists and isn't stale.
            return Promise.resolve(activeRepos)
        }
    }

    // Fetch fresh data and store it.
    return fetchActiveRepos().then(res => {
        if (res) {
            activeRepos = {
                ...activeRepos,
                timestamp: Date.now(),
                active: res.active,
                inactive: res.inactive
            }
            window.localStorage.setItem(localStorageKey, JSON.stringify(activeRepos))
            return activeRepos
        }
        return activeRepos
    })
}

export function getCurrent(): ActiveRepoResults | null {
    const data = window.localStorage.getItem(localStorageKey)
    if (data) {
        return JSON.parse(data) as ActiveRepoResults
    }
    return null
}
