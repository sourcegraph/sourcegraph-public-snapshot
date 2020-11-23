import React from 'react'
import { FileDecoration } from 'sourcegraph'
import { TreeEntryFields } from '../graphql-operations'
import classNames from 'classnames'
import { fileDecorationColorForTheme } from '../../../shared/src/api/client/services/decoration'

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

export function renderFileDecorations({
    fileDecorations,
    isLightTheme,
    isDirectory,
    isSelected,
}: {
    fileDecorations?: FileDecoration[]
    isLightTheme: boolean
    isDirectory?: boolean
    isSelected?: boolean
}): React.ReactNode {
    // Only need to check for number of decorations, other validation (like whether the decoration specifies at
    // least one of `text` or `percentage`) is done in the extension host
    if (!fileDecorations || fileDecorations.length === 0) {
        return null
    }

    return (
        <div
            className={classNames('d-flex align-items-center text-nowrap test-file-decoration-container', {
                'mr-3': isDirectory,
            })}
        >
            {fileDecorations.map(
                (fileDecoration, index) =>
                    (fileDecoration.percentage || fileDecoration.after) && (
                        <div
                            className="d-flex align-items-center"
                            // We want some margin right if this isn't the last decoration for this file
                            // eslint-disable-next-line react/forbid-dom-props
                            style={{ marginRight: index === fileDecorations.length - 1 ? 0 : 12 }}
                            key={fileDecoration.path + String(index)}
                        >
                            {fileDecoration.after && (
                                // link or span?
                                <small
                                    // eslint-disable-next-line react/forbid-dom-props
                                    style={{
                                        color: fileDecorationColorForTheme(
                                            fileDecoration.after,
                                            isLightTheme,
                                            isSelected
                                        ),
                                    }}
                                    data-tooltip={fileDecoration.after.hoverMessage}
                                    data-placement="bottom"
                                    className="text-monospace d-inline-block font-weight-normal test-file-decoration-text"
                                >
                                    {fileDecoration.after.value}
                                </small>
                            )}
                            {fileDecoration.percentage && (
                                <div
                                    className="progress rounded ml-2"
                                    // eslint-disable-next-line react/forbid-dom-props
                                    style={{ width: 24 }}
                                    data-tooltip={fileDecoration.percentage.hoverMessage}
                                    data-placement="bottom"
                                >
                                    <div
                                        className="progress-bar test-file-decoration-progress"
                                        // eslint-disable-next-line react/forbid-dom-props
                                        style={{
                                            width: `${fileDecoration.percentage.value}%`,
                                            height: 4,
                                            backgroundColor: fileDecoration.percentage.color,
                                        }}
                                        aria-valuemin={0}
                                        aria-valuemax={100}
                                        aria-valuenow={fileDecoration.percentage.value}
                                    />
                                </div>
                            )}
                        </div>
                    )
            )}
        </div>
    )
}
