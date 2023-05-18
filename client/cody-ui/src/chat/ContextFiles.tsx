import React from 'react'

import { mdiFileDocumentOutline, mdiMagnify } from '@mdi/js'

import { ContextFile } from '@sourcegraph/cody-shared/src/codebase-context/messages'
import { pluralize } from '@sourcegraph/common'

import { TranscriptAction } from './actions/TranscriptAction'

export interface FileLinkProps {
    path: string
    repoName?: string
    revision?: string
}

interface ContextItemKind {
    // The name used to describe this kind of item in the UI. Singular.
    // eg "source file"
    noun: string

    // The verb used to describe grokking this index of items
    // eg "Searched"
    verb: string

    // The object used to describe grokking this index of items
    // eg "the entire codebase for relevant files"
    object: string

    // Path to SVG icon for the retrieval operation.
    searchIcon: string

    // Path to SVG icon for each item retrieved.
    itemIcon: string

    // Tests whether the specified item is of this kind
    contains(item: ContextFile): boolean

    // Gets the React component for presenting this item
    present(item: ContextFile): JSX.Element
}

// Gets all of the known kinds of context items.
function getContextItemKinds(FileLink: React.FunctionComponent<FileLinkProps>): ContextItemKind[] {
    return [
        {
            noun: 'file',
            verb: 'Searched',
            object: 'entire codebase for relevant files',
            searchIcon: mdiMagnify,
            itemIcon: mdiFileDocumentOutline,
            contains(): boolean {
                // This kind is a catch-all which is why it is last in the list
                return true
            },
            present(file): JSX.Element {
                return <FileLink path={file.fileName} repoName={file.repoName} revision={file.revision} />
            },
        },
    ]
}

// Groups a set of context items by the kind of item.
function contextItemsByKind(kinds: ContextItemKind[], items: ContextFile[]): Map<ContextItemKind, ContextFile[]> {
    const result = new Map<ContextItemKind, ContextFile[]>()
    for (const item of items) {
        let foundKind = false
        for (const kind of kinds) {
            if (kind.contains(item)) {
                foundKind = true
                if (!result.has(kind)) {
                    result.set(kind, [])
                }
                result.get(kind)?.push(item)
                break
            }
        }
        if (!foundKind) {
            throw new Error(`context files did not find a kind to present item "${item.fileName}"`)
        }
    }
    return result
}

export const ContextFiles: React.FunctionComponent<{
    contextFiles: ContextFile[]
    fileLinkComponent: React.FunctionComponent<FileLinkProps>
    className?: string
}> = ({ contextFiles, fileLinkComponent: FileLink, className }) => {
    const kinds = getContextItemKinds(FileLink)
    const itemsByKind = new Array(...contextItemsByKind(kinds, contextFiles).entries())
    // Put more numerous items first.
    itemsByKind.sort((a, b) => b[1].length - a[1].length)

    // Construct all the steps.
    const steps = []
    for (const [kind, items] of itemsByKind) {
        steps.push({ verb: kind.verb, object: kind.object, icon: kind.searchIcon })
        steps.push(
            ...items.map(item => ({
                verb: '',
                object: kind.present(item),
                icon: kind.itemIcon,
            }))
        )
    }

    return (
        <TranscriptAction
            title={{
                verb: 'Read',
                object: `${itemsByKind.length} ${pluralize('kind', itemsByKind.length)}`,
            }}
            steps={steps}
            className={className}
        />
    )
}
