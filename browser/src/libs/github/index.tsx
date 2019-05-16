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
