import { useEffect, useMemo } from 'react'

import type * as H from 'history'
import { useLocation } from 'react-router-dom'

import type { Scalars } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

export interface UserHistoryEntry {
    repoName: string
    filePath?: string
    lastAccessed: number
}

const LAST_REPO_ACCESS_FILEPATH = 'sourcegraph-last-repo-access.timestamp'
const LOCAL_STORAGE_KEY = 'user-history'
/** Maximum number of browser history entries to persist in local storage */
const MAX_LOCAL_STORAGE_COUNT = 100

/**
 * Collects all browser history events and stores which repos/files are visited in local
 * storage. The history is used to personalize ranking in the fuzzy finder and populate
 * suggestions in the Cody context selector.
 *
 * In the future, we should consider storing this history remotely in temporary settings
 * and use them to power other features like improving ranking in the search bar
 * suggestions.
 */
export class UserHistory {
    private userID: Scalars['ID'] = 'anonymous'
    private repos: Map<string, Map<string, number>> = new Map()
    private storage = window.localStorage
    constructor(userID: Scalars['ID'] = 'anonymous') {
        this.userID = userID
        this.migrateOldEntries()
        for (const entry of this.loadEntries()) {
            this.onEntry(entry)
        }
    }
    private storageKey(userID: Scalars['ID']): string {
        return `${LOCAL_STORAGE_KEY}:${userID}`
    }
    // User history entries were previously stored in a single array in local storage
    // under the generic key `user-history`, which is not differentiated by user. The
    // first time we reinitialize UserHistory and this method runs, we take any old
    // entries and write them to the new user-differentiated key, and then delete the old
    // key. When the method runs again in the future, it will thus be a no-op.
    private migrateOldEntries(): void {
        const oldJSON = this.storage.getItem(LOCAL_STORAGE_KEY)
        if (!oldJSON) {
            return
        }
        this.storage.setItem(this.storageKey(this.userID), oldJSON)
        this.storage.removeItem(LOCAL_STORAGE_KEY)
    }
    private saveEntries(entries: UserHistoryEntry[]): void {
        entries.sort((a, b) => b.lastAccessed - a.lastAccessed)
        const truncated = entries.slice(0, MAX_LOCAL_STORAGE_COUNT)
        this.storage.setItem(this.storageKey(this.userID), JSON.stringify(truncated))
        for (let index = MAX_LOCAL_STORAGE_COUNT; index < entries.length; index++) {
            // Synchronize persisted entries with in-memory entries so that
            // reloading the page doesn't change which entries are available.
            this.deleteEntry(entries[index])
        }
    }
    public loadEntries(): UserHistoryEntry[] {
        return JSON.parse(this.storage.getItem(this.storageKey(this.userID)) ?? '[]')
    }
    private deleteEntry(entry: UserHistoryEntry): void {
        if (!entry.filePath) {
            return
        }
        const repo = this.repos.get(entry.repoName)
        if (!repo) {
            return
        }
        repo.delete(entry.filePath)
    }
    private persist(): void {
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
        this.saveEntries(entries)
    }
    private onEntry(entry: UserHistoryEntry): void {
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
    public onLocation(location: H.Location): boolean {
        try {
            const { repoName, filePath } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
            if (!repoName) {
                return false
            }
            this.onEntry({ repoName, filePath, lastAccessed: Date.now() })
            this.persist()
            return true
        } catch {
            // continue regardless of error
        }
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

/**
 * useUserHistory is a custom hook that collects browser history events for the current
 * user and stores visited repos and files in local storage.
 *
 * It takes in the user ID of the current user and whether the current page is
 * repository-related. On repository pages, it parses the location to extract the repo
 * name and file path. It then updates the history entry for that repo/file with the
 * current timestamp.
 *
 * The returned `UserHistory` instance provides methods to get the list of visited repos,
 * and lookup the last accessed timestamp for a repo or file.
 *
 * The repo history is persisted to local storage and can be used to personalize and
 * improve the search experience for the user.
 *
 * @param userID the ID of the currently-authenticated user, or undefined if the user is
 * anonymous
 * @param isRepositoryRelatedPage whether the component rendering this hook is on a page
 * that is related to a repository (e.g. a code view page) and should be tracked
 * @returns a `UserHistory` instance
 */
export function useUserHistory(userID: Scalars['ID'] | undefined, isRepositoryRelatedPage: boolean): UserHistory {
    const location = useLocation()
    const userHistory = useMemo(() => new UserHistory(userID), [userID])
    useEffect(() => {
        if (isRepositoryRelatedPage) {
            userHistory.onLocation(location)
        }
    }, [userHistory, location, isRepositoryRelatedPage])
    return userHistory
}
