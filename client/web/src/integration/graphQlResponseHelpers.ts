import { encodeURIPathComponent } from '@sourcegraph/common'
import { TreeEntriesResult } from '@sourcegraph/shared/src/graphql-operations'

import {
    BlobResult,
    FileExternalLinksResult,
    ResolveRepoRevResult,
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

export const createRepoChangesetsStatsResult = (): RepoChangesetsStatsResult => ({
    repository: {
        changesetsStats: {
            open: 2,
            merged: 4,
        },
    },
})

export const createResolveRepoRevisionResult = (treeUrl: string, oid = '1'.repeat(40)): ResolveRepoRevResult => ({
    repositoryRedirect: {
        __typename: 'Repository',
        id: `RepositoryID:${treeUrl}`,
        name: treeUrl,
        url: `/${encodeURIPathComponent(treeUrl)}`,
        externalURLs: [
            {
                url: new URL(`https://${encodeURIPathComponent(treeUrl)}`).href,
                serviceKind: ExternalServiceKind.GITHUB,
            },
        ],
        description: 'bla',
        viewerCanAdminister: false,
        defaultBranch: { displayName: 'master', abbrevName: 'master' },
        mirrorInfo: { cloneInProgress: false, cloneProgress: '', cloned: true },
        commit: {
            oid,
            tree: { url: '/' + treeUrl },
        },
    },
})

export const createFileNamesResult = (): FileNamesResult => ({
    repository: {
        id: 'repo-123',
        __typename: 'Repository',
        commit: { id: 'c0ff33', __typename: 'GitCommit', fileNames: ['README.md'] },
    },
})
