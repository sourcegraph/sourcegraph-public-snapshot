import React from 'react'
import { FileDecoration } from 'sourcegraph'
import { TreeEntryFields } from '../graphql-operations'

/** TreeEntryInfo is the information we need to render an entry in the file tree */
export interface TreeEntryInfo {
    path: string
    name: string
    isDirectory: boolean
    url: string
    submodule: TreeEntryFields['submodule']
    isSingleChild: boolean
}

export interface SingleChildGitTree extends TreeEntryInfo {
    children: SingleChildGitTree[]
}

export function scrollIntoView(element: Element, scrollRoot: Element): void {
    if (!scrollRoot.getBoundingClientRect) {
        return element.scrollIntoView()
    }

    const rootRectangle = scrollRoot.getBoundingClientRect()
    const elementRectangle = element.getBoundingClientRect()

    const elementAbove = elementRectangle.top <= rootRectangle.top + 30
    const elementBelow = elementRectangle.bottom >= rootRectangle.bottom

    if (elementAbove) {
        element.scrollIntoView(true)
    } else if (elementBelow) {
        element.scrollIntoView(false)
    }
}

export const getDomElement = (path: string): Element | null =>
    document.querySelector(`[data-tree-path='${path.replace(/'/g, "\\'")}']`)

export const treePadding = (depth: number, isTree: boolean): React.CSSProperties => ({
    marginLeft: `${depth * 12 + (isTree ? 0 : 12) + 12}px`,
    paddingRight: '16px',
})

export const maxEntries = 2500

// Utility functions to handle single-child directories:

/**
 * This function converts nested entries into a proper tree-like object. When we have single-child directories,
 * the backend responds with all entries that need to be rendered, not just the entry for that level. It is in
 * a flat list, so this function converts it to a structure like the following (assume we have a/b/c.txt):
 *
 * ```ts
 * { name: "a", ...TreeEntryInfo, children: [
 *     { name: "b", ...TreeEntryInfo, children: [
 *          {name: "c.txt", ...TreeEntryInfo, children: []}
 *     ]}
 * ]}
 * ```
 *
 * It uses the number of '/' separators to determine depth, and recursively adds entries to the `children` field.
 */
export function singleChildEntriesToGitTree(entries: TreeEntryInfo[]): SingleChildGitTree {
    const parentTree = gitTreeToTreeObject(entries[0])
    for (const [index, entry] of entries.entries()) {
        if (entry.path.split('/').length === parentTree.path.split('/').length + 1) {
            parentTree.children.push({ ...entry, children: singleChildEntriesToGitTree(entries.slice(index)).children })
        }
    }

    return parentTree
}

function gitTreeToTreeObject(entry: TreeEntryInfo): SingleChildGitTree {
    const object: SingleChildGitTree = {
        ...entry,
        children: [],
    }
    return object
}

/** Determines whether a Tree has single-child directories as children, in order to determine whether to render a SingleChildTreeLayer or TreeLayer */
export function hasSingleChild(tree: TreeEntryInfo[]): boolean {
    return tree[0]?.isSingleChild
}

// TODO(tj): impl get style by theme
export function renderFileDecorations(fileDecorations?: FileDecoration[]): React.ReactNode {
    // TODO(tj): key
    // TODO(tj): margin logic
    // early return if no decorations
    if (!fileDecorations || fileDecorations.length === 0) {
        return null
    }

    // after checking decorations, early return if no percentage or texts
    return (
        <div className="d-flex align-items-center">
            {fileDecorations.map(
                fileDecoration =>
                    (fileDecoration.percentage || fileDecoration.text) && (
                        <>
                            {fileDecoration.text && (
                                <span
                                    style={{ color: fileDecoration.text.color, fontSize: 12, textDecoration: 'none' }}
                                >
                                    {fileDecoration.text.value}
                                </span>
                            )}
                            {fileDecoration.percentage && (
                                // <progress
                                //     value={fileDecoration.percentage.value}
                                //     max="100"
                                //     style={{ width: 24, color: fileDecoration.percentage.color }}
                                // />
                                <div className="progress" style={{ width: 24, borderRadius: 4, marginLeft: 8 }}>
                                    <div
                                        className="progress-bar"
                                        // eslint-disable-next-line react/forbid-dom-props
                                        style={{
                                            width: `${fileDecoration.percentage.value}%`,
                                            height: 4,
                                            backgroundColor: fileDecoration.percentage.color,
                                        }}
                                        // TODO: aria-value
                                    />
                                </div>
                            )}
                        </>
                    )
            )}
        </div>
    )
}
