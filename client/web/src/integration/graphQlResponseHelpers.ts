import { encodeURIPathComponent } from '@sourcegraph/common'
import { TreeEntriesResult } from '@sourcegraph/shared/src/graphql-operations'

import {
    BlobResult,
    FileExternalLinksResult,
    RepositoryRedirectResult,
    ResolveRevResult,
    ExternalServiceKind,
    RepoChangesetsStatsResult,
    FileNamesResult,
} from '../graphql-operations'

export const createTreeEntriesResult = (url: string, toplevelFiles: string[]): TreeEntriesResult => ({
    repository: {
        commit: {
            tree: {
                isRoot: true,
                url,
                entries: toplevelFiles.map(name => ({
                    name,
                    path: name,
                    isDirectory: false,
                    url: `${url}/-/blob/${name}`,
                    submodule: null,
                    isSingleChild: false,
                })),
            },
        },
    },
})

export const createBlobContentResult = (
    content: string,
    html: string = `<div style="color:red">${content}<div>`
): BlobResult => ({
    repository: {
        commit: {
            file: {
                content,
                richHTML: '',
                highlight: {
                    aborted: false,
                    html,
                    lsif: '',
                },
            },
        },
    },
})

export const createFileExternalLinksResult = (
    url: string,
    serviceKind: ExternalServiceKind = ExternalServiceKind.GITHUB
): FileExternalLinksResult => ({
    repository: {
        commit: {
            file: {
                externalURLs: [{ url, serviceKind }],
            },
        },
    },
})

export const createRepositoryRedirectResult = (
    repoName: string,
    serviceKind: ExternalServiceKind = ExternalServiceKind.GITHUB
): RepositoryRedirectResult => ({
    repositoryRedirect: {
        __typename: 'Repository',
        id: `RepositoryID:${repoName}`,
        name: repoName,
        url: `/${encodeURIPathComponent(repoName)}`,
        externalURLs: [{ url: new URL(`https://${encodeURIPathComponent(repoName)}`).href, serviceKind }],
        description: 'bla',
        viewerCanAdminister: false,
        defaultBranch: { displayName: 'master', abbrevName: 'master' },
    },
})

export const createRepoChangesetsStatsResult = (): RepoChangesetsStatsResult => ({
    repository: {
        changesetsStats: {
            open: 2,
            merged: 4,
        },
    },
})

export const createResolveRevisionResult = (treeUrl: string, oid = '1'.repeat(40)): ResolveRevResult => ({
    repositoryRedirect: {
        __typename: 'Repository',
        mirrorInfo: { cloneInProgress: false, cloneProgress: '', cloned: true },
        commit: {
            oid,
            tree: { url: '/' + treeUrl },
        },
        defaultBranch: { abbrevName: 'master' },
    },
})

export const createFileNamesResult = (): FileNamesResult => ({
    repository: {
        id: 'repo-123',
        __typename: 'Repository',
        commit: { id: 'c0ff33', __typename: 'GitCommit', fileNames: ['README.md'] },
    },
})
