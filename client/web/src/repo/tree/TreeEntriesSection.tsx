import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import classNames from 'classnames'
import { Link } from '../../../../shared/src/components/Link'
import { FileDecorationsByPath } from 'sourcegraph'
import { renderFileDecorations } from '../../tree/util'

/**
 * Use a multi-column layout for tree entries when there are at least this many. See TreeEntriesSection.scss
 * for more information.
 */
const MIN_ENTRIES_FOR_COLUMN_LAYOUT = 6

const TreeEntry: React.FunctionComponent<{
    isDir: boolean
    name: string
    parentPath: string
    url: string
    fileDecorationsByPath: FileDecorationsByPath
    isColumnLayout: boolean
}> = ({ isDir, name, parentPath, url, fileDecorationsByPath, isColumnLayout }) => {
    const filePath = parentPath ? parentPath + '/' + name : name
    // TODO(tj): the parent knows the path already (entry), refactor
    const fileDecorations = fileDecorationsByPath[filePath]

    return (
        // TODO(tj): Limit width when not column layout
        <div className={classNames('d-flex align-items-center justify-content-between', { 'w-25': !isColumnLayout })}>
            <Link
                to={url}
                className={classNames(
                    'tree-entry',
                    isDir && 'font-weight-bold',
                    `test-tree-entry-${isDir ? 'directory' : 'file'}`
                )}
                title={filePath}
            >
                {name}
                {isDir && '/'}
            </Link>
            {renderFileDecorations(fileDecorations)}
        </div>
    )
}
export const TreeEntriesSection: React.FunctionComponent<{
    parentPath: string
    entries: Pick<GQL.ITreeEntry, 'name' | 'isDirectory' | 'url' | 'path'>[]
    fileDecorationsByPath: FileDecorationsByPath
}> = ({ parentPath, entries, fileDecorationsByPath }) => {
    const directChildren = entries.filter(entry => entry.path === [parentPath, entry.name].filter(Boolean).join('/'))
    if (directChildren.length === 0) {
        return null
    }

    const isColumnLayout = directChildren.length > MIN_ENTRIES_FOR_COLUMN_LAYOUT

    // TODO(tj): render file decorations for all files in parent so we know how many total file decorations exist,
    // so we can decide whether or not to render dividers

    return (
        <div className={isColumnLayout ? 'tree-entries-section--columns' : undefined}>
            {directChildren.map((entry, index) => (
                <TreeEntry
                    key={entry.name + String(index)}
                    isDir={entry.isDirectory}
                    name={entry.name}
                    parentPath={parentPath}
                    url={entry.url}
                    fileDecorationsByPath={fileDecorationsByPath}
                    isColumnLayout={isColumnLayout}
                />
            ))}
        </div>
    )
}
