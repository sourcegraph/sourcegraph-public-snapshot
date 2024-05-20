import { dirname } from 'path'

import { mapOrThrow, query } from '$lib/graphql'
import type { Scalars } from '$lib/graphql-types'
import type { TreeProvider } from '$lib/TreeView'

import { type GitCommitFieldsWithTree, type TreeEntriesVariables, TreeEntries } from './tree.gql'

const MAX_FILE_TREE_ENTRIES = 1000

/**
 * Represents the root path of the repository.
 */
export const ROOT_PATH = ''

export function isFileEntry(entry: TreeEntry): entry is Extract<TreeEntry, { __typename: 'GitBlob' }> {
    return entry.__typename === 'GitBlob' || !entry.isDirectory
}

export function fetchTreeEntries(args: TreeEntriesVariables): Promise<GitCommitFieldsWithTree> {
    return query(
        TreeEntries,
        {
            ...args,
            first: args.first ?? MAX_FILE_TREE_ENTRIES,
        }
        // mightContainPrivateInfo: true,
    ).then(
        mapOrThrow(result => {
            if (!result.data?.repository) {
                throw new Error('Unable to fetch repository information')
            }
            if (!result.data.repository.commit) {
                throw new Error('Unable to fetch commit information')
            }
            return result.data.repository.commit
        })
    )
}

export const NODE_LIMIT: unique symbol = Symbol()

type TreeRoot = NonNullable<GitCommitFieldsWithTree['tree']>
export type TreeEntry = NonNullable<GitCommitFieldsWithTree['tree']>['entries'][number]
type ExpandableFileTreeNodeValues = TreeEntry
export type FileTreeNodeValue = ExpandableFileTreeNodeValues | typeof NODE_LIMIT
export type FileTreeData = { root: TreeRoot; values: FileTreeNodeValue[] }

export async function fetchSidebarFileTree({
    repoName,
    revision,
    filePath,
}: {
    repoName: Scalars['ID']['input']
    revision: string
    filePath: string
}): Promise<FileTreeData> {
    const result = await fetchTreeEntries({
        repoName,
        revision,
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

export type FileTreeLoader = (args: { filePath: string; parent?: FileTreeProvider }) => Promise<FileTreeData>

interface FileTreeProviderArgs {
    root: NonNullable<GitCommitFieldsWithTree['tree']>
    values: FileTreeNodeValue[]
    loader: FileTreeLoader
    parent?: TreeProvider<FileTreeNodeValue>
}

export class FileTreeProvider implements TreeProvider<FileTreeNodeValue> {
    constructor(private args: FileTreeProviderArgs) {}

    public copy(args?: Partial<FileTreeProviderArgs>): FileTreeProvider {
        return new FileTreeProvider({ ...this.args, ...args })
    }

    public getRoot(): FileTreeNodeValue {
        return this.args.root
    }

    public getEntries(): FileTreeNodeValue[] {
        if (this.args.parent || this.args.root.isRoot) {
            return this.args.values
        }
        // Show an entry for traversing "up" to the parent folder
        return [this.args.root, ...this.args.values]
    }

    public async fetchChildren(entry: FileTreeNodeValue): Promise<FileTreeProvider> {
        if (!this.isExpandable(entry)) {
            // This should never happen because the caller should only call fetchChildren
            // for entries where isExpandable returns true
            throw new Error('Cannot fetch children for non-expandable tree entry')
        }

        const args = await this.args.loader({
            filePath: entry.path,
            parent: this,
        })
        return new FileTreeProvider({ ...args, loader: this.args.loader, parent: this })
    }

    public async fetchParent(): Promise<FileTreeProvider> {
        const args = await this.args.loader({
            filePath: dirname(this.args.root.path),
            parent: this,
        })
        return new FileTreeProvider({ ...args, loader: this.args.loader })
    }

    public getNodeID(entry: FileTreeNodeValue): string {
        return entry === NODE_LIMIT ? 'node-limit' : entry.path
    }

    public isExpandable(entry: FileTreeNodeValue): entry is ExpandableFileTreeNodeValues {
        return entry !== NODE_LIMIT && entry !== this.args.root && entry.isDirectory
    }

    public isSelectable(entry: FileTreeNodeValue): boolean {
        return entry !== NODE_LIMIT
    }
}
