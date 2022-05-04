import React from 'react'

import classNames from 'classnames'
import { identity } from 'lodash'
import FileDocumentOutlineIcon from 'mdi-react/FileDocumentOutlineIcon'
import FolderOutlineIcon from 'mdi-react/FolderOutlineIcon'

import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import * as GQL from '@sourcegraph/shared/src/schema'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, Icon } from '@sourcegraph/wildcard'

import { FileDecorator } from '../../tree/FileDecorator'

import styles from './TreeEntriesSection.module.scss'

/**
 * Use a multi-column layout for tree entries when there are at least this many. See TreeEntriesSection.scss
 * for more information.
 */
const MIN_ENTRIES_FOR_COLUMN_LAYOUT = 6

const TreeEntry: React.FunctionComponent<
    React.PropsWithChildren<{
        isDirectory: boolean
        name: string
        parentPath: string
        url: string
        isColumnLayout: boolean
        renderedFileDecorations: React.ReactNode
        path: string
    }>
> = ({ isDirectory, name, url, isColumnLayout, renderedFileDecorations, path }) => (
    <Link
        to={url}
        className={classNames(
            'test-page-file-decorable',
            styles.treeEntry,
            isDirectory && 'font-weight-bold',
            `test-tree-entry-${isDirectory ? 'directory' : 'file'}`,
            !isColumnLayout && styles.treeEntryNoColumns
        )}
        title={path}
        data-testid="tree-entry"
    >
        <div
            className={classNames(
                'd-flex align-items-center justify-content-between test-file-decorable-name overflow-hidden'
            )}
        >
            <span>
                <Icon className="mr-1" as={isDirectory ? FolderOutlineIcon : FileDocumentOutlineIcon} />
                {name}
                {isDirectory && '/'}
            </span>
            {renderedFileDecorations}
        </div>
    </Link>
)

interface TreeEntriesSectionProps extends ThemeProps {
    parentPath: string
    entries: Pick<GQL.ITreeEntry, 'name' | 'isDirectory' | 'url' | 'path'>[]
    fileDecorationsByPath: FileDecorationsByPath
}

export const TreeEntriesSection: React.FunctionComponent<React.PropsWithChildren<TreeEntriesSectionProps>> = ({
    parentPath,
    entries,
    fileDecorationsByPath,
    isLightTheme,
}) => {
    const directChildren = entries.filter(entry => entry.path === [parentPath, entry.name].filter(Boolean).join('/'))
    if (directChildren.length === 0) {
        return null
    }

    // Render file decorations for all files in parent so we know how many total file decorations exist
    // and can decide whether or not to render dividers
    // No need to memoize decorations, since this component should only rerender when entries change
    const renderedDecorationsByIndex = directChildren.map(entry => (
        <FileDecorator
            key={entry.path}
            // If component is not specified, or it is 'page', render it.
            fileDecorations={fileDecorationsByPath[entry.path]?.filter(decoration => decoration?.where !== 'sidebar')}
            isLightTheme={isLightTheme}
        />
    ))

    // If there are no file decorations, we want to hide column-rule.
    // TODO(tj): turn 4 iterations over directChildren in this component into 1
    const noDecorations = !directChildren
        // Return whether or not each child has decorations
        .map(entry => {
            const decorations = fileDecorationsByPath[entry.path]?.filter(decoration => decoration?.where !== 'sidebar')
            if (!decorations) {
                return false
            }

            return decorations.length > 0
        })
        // If any child has decorations, the result is true
        .find(identity)

    const isColumnLayout = directChildren.length > MIN_ENTRIES_FOR_COLUMN_LAYOUT

    return (
        <div
            className={
                isColumnLayout
                    ? classNames(
                          'pr-2',
                          styles.treeEntriesSectionColumns,
                          noDecorations && styles.treeEntriesSectionNoDecorations
                      )
                    : undefined
            }
        >
            {directChildren.map((entry, index) => (
                <TreeEntry
                    key={entry.name + String(index)}
                    parentPath={parentPath}
                    isColumnLayout={isColumnLayout}
                    renderedFileDecorations={renderedDecorationsByIndex[index]}
                    {...entry}
                />
            ))}
        </div>
    )
}
