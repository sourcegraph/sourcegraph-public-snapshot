import { dirname } from 'path'

import { query, gql } from '$lib/graphql'
import type { TreeEntriesResult, GitCommitFieldsWithTree, TreeEntriesVariables, Scalars } from '$lib/graphql-operations'
import type { TreeProvider } from '$lib/TreeView'

const MAX_FILE_TREE_ENTRIES = 1000

const treeEntriesQuery = gql`
    query TreeEntries($repoID: ID!, $commitID: String!, $filePath: String!, $first: Int) {
        node(id: $repoID) {
            __typename
            id
            ... on Repository {
                commit(rev: $commitID) {
                    ...GitCommitFieldsWithTree
                }
            }
        }
    }

    fragment GitCommitFieldsWithTree on GitCommit {
        id
        oid
        abbreviatedOID
        author {
            ...UserFields
        }
        committer {
            ...UserFields
        }
        subject

        tree(path: $filePath) {
            canonicalURL
            isRoot
            name
            path
            isDirectory
            submodule {
                commit
            }
            isSingleChild
            entries(first: $first, recursiveSingleChild: false) {
                canonicalURL
                name
                path
                isDirectory
                submodule {
                    commit
                }
                isSingleChild
            }
        }
    }

    fragment UserFields on Signature {
        person {
            name
            displayName
            avatarURL
        }
        date
    }
`

export async function fetchTreeEntries(args: TreeEntriesVariables): Promise<GitCommitFieldsWithTree> {
    const data = await query<TreeEntriesResult, TreeEntriesVariables>(
        treeEntriesQuery,
        {
            ...args,
            first: args.first ?? MAX_FILE_TREE_ENTRIES,
        }
        // mightContainPrivateInfo: true,
    )
    if (data.node?.__typename !== 'Repository' || !data.node.commit) {
        throw new Error('Unable to fetch repository information')
    }
    return data.node.commit
}

export const NODE_LIMIT: unique symbol = Symbol()

type TreeRoot = NonNullable<GitCommitFieldsWithTree['tree']>
export type TreeEntryFields = NonNullable<GitCommitFieldsWithTree['tree']>['entries'][number]
type ExpandableFileTreeNodeValues = TreeEntryFields
export type FileTreeNodeValue = ExpandableFileTreeNodeValues | typeof NODE_LIMIT

export async function fetchSidebarFileTree({
    repoID,
    commitID,
    filePath,
}: {
    repoID: Scalars['ID']
    commitID: string
    filePath: string
}): Promise<{ root: TreeRoot; values: FileTreeNodeValue[] }> {
    const result = await fetchTreeEntries({
        repoID,
        commitID,
        filePath,
        first: MAX_FILE_TREE_ENTRIES,
    })
    if (!result.tree) {
        throw new Error('Unable to fetch directory contents')
    }
    const root = result.tree
    let values: FileTreeNodeValue[] = root.entries
    if (values.length >= MAX_FILE_TREE_ENTRIES) {
        values = [...values, NODE_LIMIT]
    }
    return { root, values }
}

export type FileTreeLoader = (args: {
    repoID: Scalars['ID']
    commitID: string
    filePath: string
    parent?: FileTreeProvider
}) => Promise<FileTreeProvider>

interface FileTreeProviderArgs {
    root: NonNullable<GitCommitFieldsWithTree['tree']>
    values: FileTreeNodeValue[]
    repoID: Scalars['ID']
    commitID: string
    loader: FileTreeLoader
    parent?: TreeProvider<FileTreeNodeValue>
}

export class FileTreeProvider implements TreeProvider<FileTreeNodeValue> {
    constructor(private args: FileTreeProviderArgs) {}

    getRoot(): FileTreeNodeValue {
        return this.args.root
    }

    getRepoID(): Scalars['ID'] {
        return this.args.repoID
    }

    getEntries(): FileTreeNodeValue[] {
        if (this.args.parent || this.args.root.isRoot) {
            return this.args.values
        }
        // Show an entry for traversing "up" to the parent folder
        return [this.args.root, ...this.args.values]
    }

    async fetchChildren(entry: FileTreeNodeValue): Promise<FileTreeProvider> {
        if (!this.isExpandable(entry)) {
            // This should never happen because the caller should only call fetchChildren
            // for entries where isExpandable returns true
            throw new Error('Cannot fetch children for non-expandable tree entry')
        }

        return this.args.loader({
            repoID: this.args.repoID,
            commitID: this.args.commitID,
            filePath: entry.path,
            parent: this,
        })
    }

    async fetchParent(): Promise<FileTreeProvider> {
        const parentPath = dirname(this.args.root.path)
        return this.args.loader({
            repoID: this.args.repoID,
            commitID: this.args.commitID,
            filePath: parentPath,
        })
    }

    getNodeID(entry: FileTreeNodeValue): string {
        return entry === NODE_LIMIT ? 'node-limit' : entry.path
    }

    isExpandable(entry: FileTreeNodeValue): entry is ExpandableFileTreeNodeValues {
        return entry !== NODE_LIMIT && entry !== this.args.root && entry.isDirectory
    }

    isSelectable(entry: FileTreeNodeValue): boolean {
        return entry !== NODE_LIMIT
    }
}
