import { last, take } from 'lodash'

import { FileSpec, RawRepoSpec, RevSpec } from '../../../../shared/src/util/url'
import { commitIDFromPermalink } from '../../shared/util/dom'
import { FileInfo } from '../code_intelligence'
import { isExtension } from '../../context'

export enum GitLabPageKind {
    File,
    Commit,
    MergeRequest,
    Other,
}

/**
 * General information that can be found on any GitLab page that we care about. (i.e. has code)
 */
interface GitLabInfo extends RawRepoSpec {
    pageKind: GitLabPageKind

    owner: string
    projectName: string
}

/**
 * Information about single file pages.
 */
interface GitLabFileInfo extends RawRepoSpec, FileSpec, RevSpec {}

/**
 * Gets information about the page.
 */
export function getPageInfo(): GitLabInfo {
    const projectLink = document.querySelector<HTMLAnchorElement>('.context-header a')
    if (!projectLink) {
        throw new Error('Unable to determine project name')
    }

    const projectFullName = new URL(projectLink.href).pathname.slice(1)

    const parts = projectFullName.split('/')

    const owner = take(parts, parts.length - 1).join('/')
    const projectName = last(parts)!

    let pageKind: GitLabPageKind
    if (window.location.pathname.includes(`${owner}/${projectName}/commit`)) {
        pageKind = GitLabPageKind.Commit
    } else if (window.location.pathname.includes(`${owner}/${projectName}/merge_requests`)) {
        pageKind = GitLabPageKind.MergeRequest
    } else if (window.location.pathname.includes(`${owner}/${projectName}/blob`)) {
        pageKind = GitLabPageKind.File
    } else {
        pageKind = GitLabPageKind.Other
    }

    const hostname = isExtension ? window.location.hostname : new URL(gon.gitlab_url).hostname

    return {
        owner,
        projectName,
        rawRepoName: [hostname, owner, projectName].join('/'),
        pageKind,
    }
}

/**
 * Gets information about a file view page.
 */
export function getFilePageInfo(): GitLabFileInfo {
    const { rawRepoName, owner, projectName } = getPageInfo()

    const matches = window.location.pathname.match(new RegExp(`${owner}/${projectName}/blob/(.*?)/(.*)`))
    if (!matches) {
        throw new Error('Unable to determine revision or file path')
    }

    const rev = decodeURIComponent(matches[1])
    const filePath = decodeURIComponent(matches[2])

    return {
        rawRepoName,
        filePath,
        rev,
    }
}

const createErrorBuilder = (message: string) => (kind: string) => new Error(`${message} (${kind})`)

/**
 * Information specific to diff pages.
 */
export interface GitLabDiffInfo extends RawRepoSpec, Pick<GitLabInfo, 'owner' | 'projectName'> {
    mergeRequestID: string

    diffID?: string
    baseCommitID?: string
    baseRawRepoName: string
}

/**
 * Scrapes the DOM for the repo name and revision information.
 */
export function getDiffPageInfo(): GitLabDiffInfo {
    const { rawRepoName, owner, projectName } = getPageInfo()

    const query = new URLSearchParams(window.location.search)

    const matches = window.location.pathname.match(/merge_requests\/(.*?)\/diffs/)
    if (!matches) {
        throw new Error('Unable to determine merge request ID')
    }

    const sourceRepoLink = document.querySelector<HTMLLinkElement>('.js-source-branch a:first-child')
    if (!sourceRepoLink) {
        throw new Error('Could not find merge request source repo link')
    }
    const [baseRepoOwner, baseRepoProjectName] = new URL(sourceRepoLink.href).pathname.split('/').slice(1)
    if (!baseRepoOwner) {
        throw new Error('Could not determine MR baseRawRepoName: no baseRepoOwner')
    }
    if (!baseRepoProjectName) {
        throw new Error('Could not determine MR baseRawRepoName: no baseRepoProjectName')
    }
    const hostname = isExtension ? window.location.hostname : new URL(gon.gitlab_url).hostname
    const baseRawRepoName = `${hostname}/${owner}/${projectName}`

    return {
        baseRawRepoName,
        rawRepoName,
        owner,
        projectName,
        mergeRequestID: matches[1],
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

    const getFilePathFromElem = (elem: HTMLElement): string => {
        const filePath = elem.dataset.originalTitle || elem.dataset.title
        if (!filePath) {
            throw buildFileError('no-file-title')
        }

        return filePath
    }

    const filePathDidChange = filePathElements.length > 1
    const filePath = getFilePathFromElem(filePathElements.item(filePathDidChange ? 1 : 0))

    return {
        filePath,
        baseFilePath: filePathDidChange ? getFilePathFromElem(filePathElements.item(0)) : filePath,
    }
}

/**
 * Gets the head commit ID from the "View file @ ..." link on the code view.
 */
export function getHeadCommitIDFromCodeView(codeView: HTMLElement): FileInfo['commitID'] {
    const fileActionsLinks = codeView.querySelectorAll<HTMLLinkElement>('.file-actions a')
    for (const linkElement of fileActionsLinks) {
        const revMatch = new URL(linkElement.href).pathname.match(/blob\/(.*?)\//)
        if (revMatch) {
            return revMatch[1]
        }
    }
    throw new Error('Unable to determine head revision from code view')
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
