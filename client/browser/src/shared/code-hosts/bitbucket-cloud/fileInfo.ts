import type { BlobInfo, DiffInfo } from '../shared/codeHost'

import { getRawRepoName } from './context'
import {
    determineChangeTypeForCommit,
    determinePRChangeType,
    getCommitIDFromPermalink,
    getCommitIDsForCommit,
    getRevisionsForPR,
    parseCommitFilePaths,
    parsePRFilePaths,
} from './scrape'

export function getFileInfoFromSingleFileSourceCodeView(): BlobInfo {
    const rawRepoName = getRawRepoName()

    const revisionRegex = /src\/(.*?)\/(.*)/
    const matches = location.pathname.match(revisionRegex)
    if (!matches) {
        throw new Error('Unable to determine revision or file path')
    }
    const revision = decodeURIComponent(matches[1])
    const filePath = decodeURIComponent(matches[2])

    const commitID = getCommitIDFromPermalink()

    return {
        blob: {
            rawRepoName,
            revision,
            filePath,
            commitID: commitID ?? revision,
        },
    }
}

export function getFileInfoForPullRequest(codeView: HTMLElement): DiffInfo {
    const rawRepoName = getRawRepoName()

    const changeType = determinePRChangeType(codeView)
    const { base: baseFilePath, head: headFilePath } = parsePRFilePaths(codeView, changeType)

    const { baseRevision, headRevision } = getRevisionsForPR()

    if (changeType === 'added') {
        return {
            head: {
                rawRepoName,
                filePath: headFilePath!,
                commitID: headRevision,
                revision: headRevision,
            },
        }
    }
    if (changeType === 'removed') {
        return {
            base: {
                rawRepoName,
                filePath: baseFilePath!,
                commitID: baseRevision,
                revision: baseRevision,
            },
        }
    }
    // Modified or renamed
    return {
        head: {
            rawRepoName,
            filePath: headFilePath!,
            commitID: headRevision,
            revision: headRevision,
        },
        base: {
            rawRepoName,
            filePath: baseFilePath!,
            commitID: baseRevision,
            revision: baseRevision,
        },
    }
}

export function getFileInfoForCommit(codeView: HTMLElement): DiffInfo {
    const rawRepoName = getRawRepoName()
    const changeType = determineChangeTypeForCommit(codeView)
    const { head: headFilePath, base: baseFilePath } = parseCommitFilePaths(codeView, changeType)
    const { baseCommitID, headCommitID } = getCommitIDsForCommit()

    if (changeType === 'added') {
        return {
            head: {
                rawRepoName,
                filePath: headFilePath!,
                commitID: headCommitID,
            },
        }
    }
    if (changeType === 'removed') {
        return {
            base: {
                rawRepoName,
                filePath: baseFilePath!,
                commitID: baseCommitID,
            },
        }
    }

    return {
        head: {
            rawRepoName,
            filePath: headFilePath!,
            commitID: headCommitID,
        },
        base: {
            rawRepoName,
            filePath: baseFilePath!,
            commitID: baseCommitID,
        },
    }
}
