import { useLocalStorage } from '@sourcegraph/wildcard'
import * as H from 'history'
import { useEffect, useMemo, useRef } from 'react'
import { parseBrowserRepoURL } from '../util/url'

export interface UserHistoryEntry {
    repoName: string
    filePath?: string
    lastAccessed: number
}

const LAST_REPO_ACCESS_FILEPATH = 'sourcegraph-last-repo-access.timestamp'

/**
 * Collects all browser history events and stores which repos/files are visited
 * in local storage.  In the future, we should consider storing this history
 * remotely in temporary settings (or similar). The history is used to
 * personalize ranking in the fuzzy finder, but could theorically power other
 * features like improve ranking in the search bar suggestions.
 */
export class UserHistory {
    private repos: Map<string, Map<string, number>> = new Map()
    constructor(private setEntries: React.Dispatch<React.SetStateAction<UserHistoryEntry[]>>) {}
    public onEntry(entry: UserHistoryEntry): void {
        let repo = this.repos.get(entry.repoName)
        if (!repo) {
            repo = new Map()
            this.repos.set(entry.repoName, repo)
        }
        repo.set(LAST_REPO_ACCESS_FILEPATH, entry.lastAccessed)
        if (!entry.filePath) {
            return
        }
        repo.set(entry.filePath, entry.lastAccessed)
    }
    public persist(): void {
        const entries: UserHistoryEntry[] = []
        for (const repoName of this.repos.keys()) {
            const repoMap = this.repos.get(repoName) ?? new Map<string, number>()
            for (const filePath of repoMap.keys()) {
                const lastAccessed = repoMap.get(filePath)
                if (lastAccessed) {
                    entries.push({ repoName, filePath, lastAccessed })
                }
            }
        }
        this.setEntries(entries)
    }
    public onLocation(location: H.Location): boolean {
        try {
            const { repoName = '', filePath = '' } = parseBrowserRepoURL(
                location.pathname + location.search + location.hash
            )
            if (!repoName) {
                return false
            }
            this.onEntry({ repoName, filePath, lastAccessed: Date.now() })
            return true
        } catch (error) {
            console.log(location, error)
        } // Ignore errors
        return false
    }
    public visitedRepos(): string[] {
        return [...this.repos.keys()]
    }
    public lastAccessedRepo(repoName: string): number | undefined {
        return this.repos.get(repoName)?.get(LAST_REPO_ACCESS_FILEPATH)
    }
    public lastAccessedFilePath(repoName: string, filePath: string): number | undefined {
        return this.repos.get(repoName)?.get(filePath)
    }
}

export function useUserHistory(history: H.History, isRepositoryRelatedPage: boolean): UserHistory {
    const [entries, setEntries] = useLocalStorage<UserHistoryEntry[]>('user-history', [])
    const userHistory = useMemo(() => new UserHistory(setEntries), [setEntries])
    useEffect(() => entries.forEach(entry => userHistory.onEntry(entry)), [])
    if (isRepositoryRelatedPage) {
        userHistory.onLocation(history.location)
    }
    return userHistory
}
