import { last } from 'lodash'

import { FileInfo } from '../code_intelligence'

export enum GitLabPageKind {
    File,
    Commit,
    MergeRequest,
}

/**
 * General information that can be found on any GitLab page that we care about. (i.e. has code)
 */
export interface GitLabInfo {
    pageKind: GitLabPageKind

    owner: string
    repoName: string

    repoPath: string

    /**
     * The parts of the URL following the repo name.
     */
    urlParts: string[]
}

/**
 * Information about single file pages.
 */
export interface GitLabFileInfo extends Pick<GitLabInfo, 'repoPath'> {
    filePath: string
    rev: string
}

/**
 * Gets information about the page.
 */
export function getPageInfo(): GitLabInfo {
    const host = window.location.hostname

    const parts = window.location.pathname.slice(1).split('/')

    const owner = parts[0]
    const repoName = parts[1]

    let pageKind: GitLabPageKind
    if (window.location.pathname.includes(`${owner}/${repoName}/commit`)) {
        pageKind = GitLabPageKind.Commit
    } else if (window.location.pathname.includes(`${owner}/${repoName}/merge_requests`)) {
        pageKind = GitLabPageKind.MergeRequest
    } else {
        pageKind = GitLabPageKind.File
    }

    return {
        owner,
        repoName,
        repoPath: [host, owner, repoName].join('/'),
        urlParts: parts.slice(2),
        pageKind,
    }
}

/**
 * Gets information about a file view page.
 */
export function getFilePageInfo(): GitLabFileInfo {
    const { repoPath } = getPageInfo()

    const parts = window.location.pathname.slice(1).split('/')

    const rev = parts[3]
    const filePath = parts.slice(4).join('/')

    return {
        repoPath,
        filePath,
        rev,
    }
}

const createErrorBuilder = (message: string) => (kind: string) => new Error(`${message} (${kind})`)

/**
 * Information specific to diff pages.
 */
export interface GitLabDiffInfo extends Pick<GitLabFileInfo, 'repoPath'>, Pick<GitLabInfo, 'owner' | 'repoName'> {
    mergeRequestID: string

    diffID?: string
    baseCommitID?: string
}

/**
 * Scrapes the DOM for the repo path and revision information.
 */
export function getDiffPageInfo(): GitLabDiffInfo {
    const { repoPath, owner, repoName, urlParts } = getPageInfo()

    const query = new URLSearchParams(window.location.search)

    return {
        repoPath,
        owner,
        repoName,
        mergeRequestID: urlParts[1],
        diffID: query.get('diff_id') || undefined,
        baseCommitID: query.get('start_sha') || undefined,
    }
}

const buildFileError = createErrorBuilder('Unable to file information')

/**
 * Finds the file paths from the code view. If the name has changed, it'll return the base and head file paths.
 */
export function getFilePathsFromCodeView(codeView: HTMLElement): Pick<FileInfo, 'filePath' | 'baseFilePath'> {
    const filePathElements = codeView.querySelectorAll<HTMLElement>('.file-title-name')
    if (filePathElements.length === 0) {
        throw buildFileError('no-file-title-element')
    }

    const getFilePathFromElem = (elem: HTMLElement) => {
        const filePath = elem.dataset.originalTitle || elem.dataset.title
        if (!filePath) {
            throw buildFileError('no-file-title')
        }

        return filePath
    }

    const filePathDidChange = filePathElements.length > 1

    return {
        filePath: getFilePathFromElem(filePathElements.item(filePathDidChange ? 1 : 0)),
        baseFilePath: filePathDidChange ? getFilePathFromElem(filePathElements.item(0)) : undefined,
    }
}

/**
 * Gets the head commit ID from the "View file @ ..." link on the code view.
 */
export function getHeadCommitIDFromCodeView(codeView: HTMLElement): FileInfo['commitID'] {
    const commitSHA = codeView.querySelector<HTMLElement>('.file-actions .commit-sha')
    if (!commitSHA) {
        throw buildFileError('no-commit-sha')
    }

    const commitAnchor = commitSHA.closest('a')! as HTMLAnchorElement
    const hrefParts = new URL(commitAnchor.href).pathname.slice(1).split('/')

    return hrefParts[3]
}

interface GitLabCommitPageInfo extends Pick<GitLabFileInfo, 'repoPath'>, Pick<GitLabInfo, 'owner' | 'repoName'> {
    commitID: FileInfo['commitID']
}

/**
 * Get the commit from the URL.
 */
export function getCommitPageInfo(): GitLabCommitPageInfo {
    const { repoPath, owner, repoName } = getPageInfo()

    return {
        repoPath,
        owner,
        repoName,
        commitID: last(window.location.pathname.split('/'))!,
    }
}
