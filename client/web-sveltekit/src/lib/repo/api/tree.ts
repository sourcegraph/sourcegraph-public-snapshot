import { mdiFileOutline, mdiFolderOpenOutline, mdiFolderOutline } from '@mdi/js'
import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, isErrorLike, memoizeObservable } from '$lib/common'
import type { TreeEntriesResult, GitCommitFieldsWithTree, TreeFields, TreeEntryFields } from '$lib/graphql-operations'
import { gql } from '$lib/http-client'
import { makeRepoURI, type AbsoluteRepoFile } from '$lib/shared'
import { requestGraphQL } from '$lib/web'

import { type TreeProvider, DummyTreeProvider } from '../domain/tree'

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
                    path
                    isRoot
                    url
                    entries(first: $first, recursiveSingleChild: true) {
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
    private openEntries = new Set<TreeEntryFields>()

    constructor(private args: FileTreeProviderArgs) {}
    getDisplayName(entry: TreeEntryFields): string {
        return entry.name
    }
    canOpen(entry: TreeEntryFields): boolean {
        return entry.isDirectory
    }
    getSVGIconPath(entry: TreeEntryFields, open: boolean): string {
        return entry.isDirectory ? (open ? mdiFolderOpenOutline : mdiFolderOutline) : mdiFileOutline
    }
    getEntries(): TreeEntryFields[] {
        return this.args.tree.entries
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
    getURL(entry: TreeEntryFields): string | null {
        return entry.url
    }
    getKey(entry: TreeEntryFields): string {
        return entry.path
    }
    markOpen(entry: TreeEntryFields, open: boolean): void {
        if (this.args.parent) {
            this.args.parent.markOpen(entry, open)
        }
        if (open) {
            this.openEntries.add(entry)
        } else {
            this.openEntries.delete(entry)
        }
    }
    isOpen(entry: TreeEntryFields): boolean {
        return this.args.parent ? this.args.parent.isOpen(entry) : this.openEntries.has(entry)
    }
}
