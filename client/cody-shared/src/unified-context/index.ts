import { ContextSearchOptions } from '../codebase-context'

export interface UnifiedContextFetcherResult {
    filePath: string
    content: string
    startLine: number
    endLine: number
    repoName: string
    revision: string
    commit?: {
        id: string
        oid: string
        date: string
        author: string
        subject: string
    }
    owner?: {
        reason?:
            | 'CodeOwnersFileEntry'
            | 'AssignedOwner'
            | 'RecentViewOwnershipSignal'
            | 'RecentContributorOwnershipSignal'
        type: 'Person' | 'Team'
        name: string
    }
}

export interface UnifiedContextFetcher {
    getContext(query: string, options: ContextSearchOptions): Promise<UnifiedContextFetcherResult[] | Error>
}
