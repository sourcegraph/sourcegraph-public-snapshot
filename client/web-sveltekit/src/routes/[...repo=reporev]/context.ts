import { getContext, setContext } from 'svelte'
import type { Writable } from 'svelte/store'

export interface RepositoryPageContext {
    revision?: string
    filePath?: string
    directoryPath?: string
    fileLanguage?: string
}

const REPOSITORY_CONTEXT_KEY = {}

export function setRepositoryPageContext(store: Writable<RepositoryPageContext>): void {
    setContext(REPOSITORY_CONTEXT_KEY, store)
}

export function getRepositoryPageContext(): Writable<RepositoryPageContext> {
    return getContext(REPOSITORY_CONTEXT_KEY)
}
