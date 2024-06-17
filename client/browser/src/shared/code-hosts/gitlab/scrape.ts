import { last, take } from 'lodash'
import { BehaviorSubject } from 'rxjs'

import type { FileSpec, RawRepoSpec, RepoSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { commitIDFromPermalink } from '../../util/dom'
import type { FileInfo } from '../shared/codeHost'

export enum GitLabPageKind {
    File,
    Commit,
    MergeRequest,
    Other,
}

/**
 * General information that can be found on any GitLab page that we care about. (i.e. has code)
 */
export interface GitLabInfo extends RawRepoSpec, RepoSpec {
    pageKind: GitLabPageKind

    owner: string
    projectName: string
}

/**
 * Information about single file pages.
 */
interface GitLabFileInfo extends RawRepoSpec, FileSpec, RevisionSpec {}

export const getPageKindFromPathName = (owner: string, projectName: string, pathname: string): GitLabPageKind => {
    const pageKindMatch = pathname.match(new RegExp(`^/${owner}/${projectName}(/-)?/(commit|merge_requests|blob)/`))
    if (!pageKindMatch) {
        return GitLabPageKind.Other
    }
    switch (pageKindMatch[2]) {
        case 'commit': {
            return GitLabPageKind.Commit
        }
        case 'merge_requests': {
            return GitLabPageKind.MergeRequest
        }
        case 'blob': {
            return GitLabPageKind.File
        }
        default: {
            return GitLabPageKind.Other
        }
    }
}

/**
 * Gets repo URL from on GitLab.
 */
export const getGitlabRepoURL = (): string => {
    const projectLink = document.querySelector<HTMLAnchorElement>('.context-header a, .shortcuts-project')

    if (!projectLink) {
        throw new Error('Unable to determine project name')
    }
    return projectLink.href // e.g. 'https://gitlab.com/SourcegraphCody/jsonrpc2'
}

const parseFullProjectName = (fullProjectName: string): { owner: string; projectName: string } => {
    const parts = fullProjectName.split('/')
    const owner = take(parts, parts.length - 1).join('/')
    const projectName = last(parts)!
    return { owner, projectName }
}

const parseGitLabRepoURL = (): { hostname: string; projectFullName: string; owner: string; projectName: string } => {
    const url = new URL(getGitlabRepoURL())
    const projectFullName = url.pathname.slice(1) // e.g. '/sourcegraph/jsonrpc2' -> 'sourcegraph/jsonrpc2'
    const { owner, projectName } = parseFullProjectName(projectFullName)
    return { hostname: url.hostname, projectFullName, owner, projectName }
}

/**
 * Subject to store repo name on the Sourcegraph instance (e.g. 'gitlab.com/SourcegraphCody/jsonrpc2').
 * It may be different from the repo name on the code host because of the name transformations applied
 * (see {@link https://sourcegraph.com/docs/admin/code_hosts/gitlab#nameTransformations}).
 * Set in `gitlabCodeHost.prepareCodeHost` method.
 */
export const repoNameOnSourcegraph = new BehaviorSubject<string>('')

/**
 * Gets information about the page.
 */
export function getPageInfo(): GitLabInfo {
    const {
        projectFullName: projectFullNameOnGitLab,
        owner: ownerOnGitLab,
        projectName: projectNameOnGitLab,
    } = parseGitLabRepoURL()

    /**
     * Get the repository name that we can use to interact with the Sourcegraph API.
     * It is possible that this differs from `projectNameOnGitLab` if the Sourcegraph instance
     * has `nameTransformations` or `repositoryPathPattern` set.
     */
    const sourcegraphCompatibleProjectName = repoNameOnSourcegraph.value
    const pageKind = getPageKindFromPathName(ownerOnGitLab, projectNameOnGitLab, window.location.pathname)

    return {
        owner: ownerOnGitLab,
        projectName: projectNameOnGitLab,
        rawRepoName: sourcegraphCompatibleProjectName,
        repoName: projectFullNameOnGitLab, // original (untransformed) repo name to be use in GitLab API calls
        pageKind,
    }
}

/**
 * Gets information about a file view page.
 */
export function getFilePageInfo(): GitLabFileInfo {
    const { rawRepoName } = getPageInfo()
    const matches = window.location.pathname.match(/\/blob\/(.*?)\/(.*)/)
    if (!matches) {
        throw new Error('Unable to determine revision or file path')
    }

    const revision = decodeURIComponent(matches[1])
    const filePath = decodeURIComponent(matches[2])
    return {
        rawRepoName,
        filePath,
        revision,
    }
}

/**
 * Finds the merge request ID from the URL.
 */
export const getMergeRequestID = (): string => {
    let matches = window.location.pathname.match(/merge_requests\/(.*?)\/diffs/)

    // If /diffs hasn't been added to the path as a result of clicking the "Changes" tab (a known GitLab bug),
    // check if the "Changes" tab is active. If so, try to find the merge request ID again.
    if (!matches && !!document.querySelector('.diffs-tab.active')) {
        // Matches with and without trailing slash (merge_requests/151 or merge_requests/151/)
        matches = window.location.pathname.match(/merge_requests\/(.*?)((\/)|$)/)
    }
    if (!matches) {
        throw new Error('Unable to determine merge request ID')
    }
    return matches[1]
}

/**
 * Finds the diff ID, if any, from the URL.
 * The diff ID represents a specific revision in a merge request.
 */
export const getDiffID = (): string | undefined => {
    const parameters = new URLSearchParams(window.location.search)
    return parameters.get('diff_id') ?? undefined
}

const getFilePathFromElement = (element: HTMLElement): string => {
    const filePath = element.dataset.originalTitle || element.dataset.title || element.title
    if (!filePath) {
        throw new Error('Unable to get file paths from code view: no file title')
    }

    // Deleted files in MRs include " deleted" after the filepath of deleted files
    if (filePath.endsWith(' deleted')) {
        return filePath.slice(0, -8)
    }

    return filePath
}

/**
 * Finds the file paths from the code view. If the name has changed, it'll return the base and head file paths.
 */
export function getFilePathsFromCodeView(codeView: HTMLElement): { headFilePath: string; baseFilePath: string } {
    const filePathElements = codeView.querySelectorAll<HTMLElement>('.file-title-name')
    if (filePathElements.length === 0) {
        throw new Error('Unable to get file paths from code view: no .file-title.name')
    }

    const filePathDidChange = filePathElements.length > 1
    const filePath = getFilePathFromElement(filePathElements.item(filePathDidChange ? 1 : 0))

    return {
        headFilePath: filePath,
        baseFilePath: filePathDidChange ? getFilePathFromElement(filePathElements.item(0)) : filePath,
    }
}

interface GitLabCommitPageInfo extends RawRepoSpec, Pick<GitLabInfo, 'owner' | 'projectName'> {
    commitID: FileInfo['commitID']
}

/**
 * Get the commit from the URL.
 */
export function getCommitPageInfo(): GitLabCommitPageInfo {
    const { rawRepoName, owner, projectName } = getPageInfo()

    return {
        rawRepoName,
        owner,
        projectName,
        commitID: last(window.location.pathname.split('/'))!,
    }
}

/**
 * Get the commit ID from the permalink element on the page.
 */
export function getCommitIDFromPermalink(): string {
    return commitIDFromPermalink({
        selector: '.js-data-file-blob-permalink-url',
        hrefRegex: new RegExp('^/.*?/.*?/blob/([0-9a-f]{40})/'),
    })
}
