import { useEffect, useMemo } from 'react'

import * as H from 'history'
import { useLocation } from 'react-router-dom'

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
 * Collects all browser history events and stores which repos/files are visited
 * in local storage.  In the future, we should consider storing this history
 * remotely in temporary settings (or similar). The history is used to
 * personalize ranking in the fuzzy finder, but could theorically power other
 * features like improve ranking in the search bar suggestions.
 */
export class UserHistory {
    private repos: Map<string, Map<string, number>> = new Map()
    private storage = window.localStorage
    constructor() {
        for (const entry of this.loadEntries()) {
            this.onEntry(entry)
        }
    }
    private saveEntries(entries: UserHistoryEntry[]): void {
        entries.sort((a, b) => b.lastAccessed - a.lastAccessed)
        const truncated = entries.slice(0, MAX_LOCAL_STORAGE_COUNT)
        this.storage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(truncated))
        for (let index = MAX_LOCAL_STORAGE_COUNT; index < entries.length; index++) {
            // Synchronize persisted entries with in-memory entries so that
            // reloading the page doesn't change which entries are available.
            this.deleteEntry(entries[index])
        }
    }
    private loadEntries(): UserHistoryEntry[] {
        return JSON.parse(this.storage.getItem(LOCAL_STORAGE_KEY) ?? '[]')
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

export function useUserHistory(isRepositoryRelatedPage: boolean): UserHistory {
    const location = useLocation()
    const userHistory = useMemo(() => new UserHistory(), [])
    useEffect(() => {
        if (isRepositoryRelatedPage) {
            userHistory.onLocation(location)
        }
    }, [userHistory, location, isRepositoryRelatedPage])
    return userHistory
}
