import { createAggregateError } from '../../../../../shared/src/util/errors'
import { DiffResolvedRevSpec } from '../../shared/repo'
import { FileInfo } from '../code_intelligence'

export interface BitbucketRepoInfo {
    repoSlug: string
    project: string
}

/** Regular expression to check if a string looks like a git commit SHA1 */
const COMMIT_ID_REGEXP = /^[a-f0-9]{40}$/

const LINK_SELECTORS = ['a.raw-view-link', 'a.source-view-link', 'a.mode-source']

/**
 * Attempts to parse the file info from a link element contained in the code view.
 * Depending on the configuration of the page, this can be a link to the raw file,
 * or to the original source view, so a few different selectors are tried.
 *
 * The href of these links contains:
 * - project name
 * - repo name
 * - file path
 * - rev (through the query parameter 'at')
 */
const getFileInfoFromLink = (
    codeView: HTMLElement
): Pick<FileInfo, 'repoName' | 'filePath' | 'rev'> & { commitID?: string } & BitbucketRepoInfo => {
    const errors: Error[] = []
    for (const selector of LINK_SELECTORS) {
        try {
            const linkElement = codeView.querySelector(selector) as HTMLLinkElement | null
            if (!linkElement) {
                throw new Error(`Could not find selector ${selector} in code view`)
            }
            const url = new URL(linkElement.href)
            const host = window.location.hostname
            const path = url.pathname

            // Looks like /projects/<project>/repos/<repo>/(browse|raw)/<file path>?at=<rev>
            const pathMatch = path.match(/\/projects\/(.*?)\/repos\/(.*?)\/(?:browse|raw)\/(.*)$/)
            if (!pathMatch) {
                throw new Error(`Path of link matching selector ${selector} did not match path regex: ${path}`)
            }

            const [, project, repoSlug, filePath] = pathMatch

            // Looks like 'refs/heads/<rev>'
            const at = url.searchParams.get('at')
            if (!at) {
                throw new Error(`href of link matching selector ${selector} did not have 'at' search param: ${url}`)
            }

            const atMatch = at.match(/refs\/heads\/(.*?)$/)

            const rev = atMatch ? atMatch[1] : at

            const commitID = COMMIT_ID_REGEXP.test(rev) ? rev : undefined

            return {
                repoName: [host, project, repoSlug].join('/'),
                filePath,
                rev,
                commitID,
                project,
                repoSlug,
            }
        } catch (err) {
            errors.push(err)
            continue
        }
    }
    throw createAggregateError(errors)
}

/**
 * Attempts to retreive the commitid from a link to the commit,
 * found on single file "diff to previous" views.
 */
const getCommitIDFromLink = (): string => {
    const commitLink = document.querySelector<HTMLElement>('a.commitid')
    if (!commitLink) {
        throw new Error('No element found matching a.commitid')
    }
    const commitID = commitLink.dataset.commitid
    if (!commitID) {
        throw new Error('Element matching a.commitid has no data-commitid')
    }
    return commitID
}

export const getFileInfoFromCodeView = (
    codeView: HTMLElement
): BitbucketRepoInfo & Pick<FileInfo, 'repoName' | 'filePath' | 'rev' | 'commitID'> => {
    const { repoName, filePath, rev, project, repoSlug, commitID } = getFileInfoFromLink(codeView)
    return {
        repoName,
        filePath,
        rev,
        commitID: commitID || getCommitIDFromLink(),
        project,
        repoSlug,
    }
}

const getFileInfoFromFilePathLink = (
    codeView: HTMLElement
): Pick<FileInfo, 'repoName' | 'filePath'> & Partial<Pick<FileInfo, 'commitID'>> & BitbucketRepoInfo => {
    const rawViewLink = codeView.querySelector<HTMLAnchorElement>('.breadcrumbs a.stub')
    if (!rawViewLink) {
        throw new Error('could not find raw view link for code view (.breadcrumbs a.stub)')
    }

    const url = new URL(rawViewLink.href)

    const host = window.location.hostname

    const path = url.pathname

    const pathMatch = path.match(/\/projects\/(.*?)\/repos\/(.*?)\/pull-requests\/(\d*)\//)
    if (!pathMatch) {
        throw new Error('Unable to parse file information')
    }

    const project = pathMatch[1]
    const repoSlug = pathMatch[2]

    const commitMatch = path.match(/\/commits\/(.*?)$/)

    const commitID = commitMatch ? commitMatch[1] : undefined

    let filePath = url.hash.replace(/^#/, '')
    filePath = filePath.replace(/\?.*$/, '')

    return {
        repoName: [host, project, repoSlug].join('/'),
        filePath,
        commitID,
        project,
        repoSlug,
    }
}

export const getDiffFileInfoFromCodeView = (
    codeView: HTMLElement
): BitbucketRepoInfo & {
    // FileInfo.commitID is required but we won't always have it from the page DOM.
    commitID?: string
} & Pick<FileInfo, 'repoName' | 'filePath'> => {
    let repoName: string
    let filePath: string
    let project: string
    let repoSlug: string
    let commitID: string | undefined

    try {
        ;({ repoName, filePath, project, repoSlug, commitID } = getFileInfoFromLink(codeView))
    } catch (e) {
        ;({ repoName, filePath, project, repoSlug, commitID } = getFileInfoFromFilePathLink(codeView))
    }

    if (!commitID) {
        try {
            commitID = getCommitIDFromLink()
        } catch (err) {
            console.error('unable to get commitID from link', err)
        }
    }

    return {
        repoName,
        filePath,
        commitID,
        project,
        repoSlug,
    }
}

export function getPRIDFromPathName(): number {
    const prIDMatch = window.location.pathname.match(/pull-requests\/(\d*?)\/(diff|overview|commits)/)
    if (!prIDMatch) {
        throw new Error(`Could not parse PR ID from pathname: ${window.location.pathname}`)
    }
    return parseInt(prIDMatch[1], 10)
}

export function getResolvedDiffFromBranchComparePage(): DiffResolvedRevSpec {
    const headCommitElement = document.querySelector('#branch-compare .source-selector a.commitid[data-commitid]')
    const baseCommitElement = document.querySelector('#branch-compare .target-selector a.commitid[data-commitid]')
    if (!headCommitElement || !baseCommitElement) {
        throw new Error('Could not resolve Bitbucket compare diff spec')
    }
    return {
        headCommitID: headCommitElement.getAttribute('data-commitid')!,
        baseCommitID: baseCommitElement.getAttribute('data-commitid')!,
    }
}
