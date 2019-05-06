import { ParsedRepoURI } from '../../../../shared/src/util/url'

export interface GitHubURL extends ParsedRepoURI {
    user?: string
    ghRepoName?: string
    isDelta?: boolean
    isPullRequest?: boolean
    isCommit?: boolean
    isCodePage?: boolean
    isCompare?: boolean
}

export interface GitHubBlobUrl {
    mode: GitHubMode
    owner: string
    ghRepoName: string
    revAndPath: string
    lineNumber: string | undefined
    rev: string
    filePath: string
}

export interface GitHubPullUrl {
    mode: GitHubMode
    owner: string
    ghRepoName: string
    view: string
    rev: string
    id: number
    filePath?: string
}

export interface GitHubRepositoryUrl {
    mode: GitHubMode
    owner: string
    ghRepoName: string
    rev?: string
    filePath?: string
}

export enum GitHubMode {
    Blob,
    Commit,
    PullRequest,
    Repository,
}
