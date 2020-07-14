import {
    TreeEntriesResult,
    BlobResult,
    FileExternalLinksResult,
    RepositoryRedirectResult,
    ResolveRevResult,
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
                },
            },
        },
    },
})

export const createFileExternalLinksResult = (
    url: string,
    serviceType: string = 'github'
): FileExternalLinksResult => ({
    repository: {
        commit: {
            file: {
                externalURLs: [{ url, serviceType }],
            },
        },
    },
})

export const createRepositoryRedirectResult = (
    repoUrl: string,
    serviceType: string = 'github'
): RepositoryRedirectResult => ({
    repositoryRedirect: {
        __typename: 'Repository',
        id: 'UmVwb3NpdG9yeTo0MDk1Mzg=', // TODO is this ok to be hardcoded?
        name: repoUrl,
        url: `/${repoUrl}`,
        externalURLs: [{ url: `https://${repoUrl}`, serviceType }],
        description: 'bla',
        viewerCanAdminister: false,
        defaultBranch: { displayName: 'master' },
    },
})

export const createResolveRevisionResult = (treeUrl: string): ResolveRevResult => ({
    repositoryRedirect: {
        __typename: 'Repository',
        mirrorInfo: { cloneInProgress: false, cloneProgress: '', cloned: true },
        commit: {
            oid: '15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81', // TODO is this ok to be hardcoded?
            tree: { url: treeUrl },
        },
        defaultBranch: { abbrevName: 'master' },
    },
})
