import React from 'react'

import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'
import classNames from 'classnames'

import { TreeEntryFields } from '@sourcegraph/shared/src/graphql-operations'
import { Link, Icon } from '@sourcegraph/wildcard'

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
        path: string
    }>
> = ({ isDirectory, name, url, isColumnLayout, path }) => (
    <li>
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
                    <Icon
                        className="mr-1"
                        svgPath={isDirectory ? mdiFolderOutline : mdiFileDocumentOutline}
                        aria-hidden={true}
                    />
                    {name}
                    {isDirectory && '/'}
                </span>
            </div>
        </Link>
    </li>
)

interface TreeEntriesSectionProps {
    parentPath: string
    entries: Pick<TreeEntryFields, 'name' | 'isDirectory' | 'url' | 'path'>[]
}

export const TreeEntriesSection: React.FunctionComponent<React.PropsWithChildren<TreeEntriesSectionProps>> = ({
    parentPath,
    entries,
}) => {
    const directChildren = entries.filter(entry => entry.path === [parentPath, entry.name].filter(Boolean).join('/'))
    if (directChildren.length === 0) {
        return null
    }

    const isColumnLayout = directChildren.length > MIN_ENTRIES_FOR_COLUMN_LAYOUT

    return (
        <ul
            className={classNames(
                'list-unstyled',
                isColumnLayout && classNames('pr-2', styles.treeEntriesSectionColumns)
            )}
        >
            {directChildren.map((entry, index) => (
                <TreeEntry
                    key={entry.name + String(index)}
                    parentPath={parentPath}
                    isColumnLayout={isColumnLayout}
                    {...entry}
                />
            ))}
        </ul>
    )
}
