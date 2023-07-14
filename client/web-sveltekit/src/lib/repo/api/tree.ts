import { dirname } from 'path'

import { memoize } from 'lodash'
import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '$lib/common'
import type { TreeEntriesResult, GitCommitFieldsWithTree, TreeFields, TreeEntryFields } from '$lib/graphql-operations'
import { gql } from '$lib/http-client'
import { makeRepoURI, type AbsoluteRepoFile } from '$lib/shared'
import type { TreeProvider } from '$lib/TreeView'
import { requestGraphQL } from '$lib/web'

export const fetchTreeEntries = memoizeObservable(
    (args: AbsoluteRepoFile & { first?: number }): Observable<GitCommitFieldsWithTree> =>
        requestGraphQL<TreeEntriesResult>(
            gql`
                query TreeEntries(
                    $repoName: String!
                    $revision: String!
                    $commitID: String!
                    $filePath: String!
                    $first: Int
                ) {
                    repository(name: $repoName) {
                        id
                        commit(rev: $commitID, inputRevspec: $revision) {
                            ...GitCommitFieldsWithTree
                        }
                    }
                }

                fragment GitCommitFieldsWithTree on GitCommit {
                    oid
                    abbreviatedOID
                    url
                    author {
                        ...UserFields
                    }
                    committer {
                        ...UserFields
                    }
                    subject

                    tree(path: $filePath) {
                        ...TreeFields
                    }
                }
                fragment TreeFields on GitTree {
                    ...TreeEntryFields
                    entries(first: $first, recursiveSingleChild: false) {
                        ...TreeEntryFields
                    }
                }
                fragment TreeEntryFields on TreeEntry {
                    name
                    path
                    isDirectory
                    url
                    submodule {
                        url
                        commit
                    }
                    isSingleChild
                    ...GitTreeEntry
                }
                fragment GitTreeEntry on GitTree {
                    isRoot
                }

                fragment UserFields on Signature {
                    person {
                        name
                        displayName
                        avatarURL
                    }
                    date
                }
            `,
            args
            //mightContainPrivateInfo: true,
        ).pipe(
            map(({ data, errors }) => {
                if (errors || !data?.repository?.commit?.tree) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit
            })
        ),
    ({ first, ...args }) => `${makeRepoURI(args)}:first-${String(first)}`
)

const MAX_FILE_TREE_ENTRIES = 1000
export const NODE_LIMIT: unique symbol = Symbol()
type ExpandableFileTreeNodeValues = TreeEntryFields
type FileTreeNodeValue = ExpandableFileTreeNodeValues | typeof NODE_LIMIT

export const fetchSidebarFileTree = memoize(
    async ({
        repoName,
        commitID,
        revision,
        filePath,
    }: {
        repoName: string
        commitID: string
        revision: string
        filePath: string
    }): Promise<{ root: TreeFields; values: FileTreeNodeValue[] }> => {
        const result = await fetchTreeEntries({
            repoName,
            commitID,
            revision,
            filePath,
            first: MAX_FILE_TREE_ENTRIES,
        }).toPromise()
        if (!result.tree) {
            throw new Error('Unable to fetch directory contents')
        }
        const root = result.tree
        let values: FileTreeNodeValue[] = root.entries
        if (values.length >= MAX_FILE_TREE_ENTRIES) {
            values = [...values, NODE_LIMIT]
        }

        return { root, values }
    },
    args => `${makeRepoURI(args)}:first-${String(MAX_FILE_TREE_ENTRIES)}`
)

export interface FileTreeLoader {
    (args: {
        repoName: string
        commitID: string
        revision: string
        filePath: string
        parent?: FileTreeProvider
    }): Promise<FileTreeProvider>
}

interface FileTreeProviderArgs {
    root: TreeFields
    values: FileTreeNodeValue[]
    repoName: string
    commitID: string
    revision: string
    loader: FileTreeLoader
    parent?: TreeProvider<FileTreeNodeValue>
}

export class FileTreeProvider implements TreeProvider<FileTreeNodeValue> {
    constructor(private args: FileTreeProviderArgs) {}

    getRoot(): FileTreeNodeValue {
        return this.args.root
    }

    getRepoName(): string {
        return this.args.repoName
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
            repoName: this.args.repoName,
            commitID: this.args.commitID,
            revision: this.args.revision,
            filePath: entry.path,
            parent: this,
        })
    }

    async fetchParent(): Promise<FileTreeProvider> {
        const parentPath = dirname(this.args.root.path)
        return this.args.loader({
            repoName: this.args.repoName,
            commitID: this.args.commitID,
            revision: this.args.revision,
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
