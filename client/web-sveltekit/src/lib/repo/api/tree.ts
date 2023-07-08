import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { writable, readonly, type Writable } from 'svelte/store'
import type { Readable } from 'svelte/store'

import { createAggregateError, isErrorLike, memoizeObservable } from '$lib/common'
import type { TreeEntriesResult, GitCommitFieldsWithTree, TreeFields, TreeEntryFields } from '$lib/graphql-operations'
import { gql } from '$lib/http-client'
import { makeRepoURI, type AbsoluteRepoFile } from '$lib/shared'
import { DummyTreeProvider, type NodeState, type TreeProvider } from '$lib/TreeView'
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

interface FileTreeProviderArgs {
    tree: TreeFields
    repoName: string
    commitID: string
    revision: string
    parent?: FileTreeProvider
}

export class FileTreeProvider implements TreeProvider<TreeEntryFields> {
    constructor(private args: FileTreeProviderArgs) {}

    getRoot(): TreeEntryFields {
        return this.args.tree
    }

    getEntries(): TreeEntryFields[] {
        if (this.args.parent || this.args.tree.isRoot) {
            return this.args.tree.entries
        }
        return [this.args.tree, ...this.args.tree.entries]
    }

    async fetchChildren(entry: TreeEntryFields): Promise<TreeProvider<TreeEntryFields>> {
        const result = await fetchTreeEntries({
            repoName: this.args.repoName,
            commitID: this.args.commitID,
            revision: this.args.revision ?? '',
            filePath: entry.path,
            first: 2500,
        }).toPromise()

        if (isErrorLike(result) || !result.tree) {
            return new DummyTreeProvider()
        }
        return new FileTreeProvider({
            ...this.args,
            tree: result.tree,
            parent: this,
        })
    }
    getKey(entry: TreeEntryFields): string {
        return entry.path
    }
    isExpandable(entry: TreeEntryFields): boolean {
        return entry !== this.args.tree && entry.isDirectory
    }
}
