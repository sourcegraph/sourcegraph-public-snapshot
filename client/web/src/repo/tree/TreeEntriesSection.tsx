import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import classNames from 'classnames'
import { Link } from '../../../../shared/src/components/Link'
import { FileDecorationsByPath } from 'sourcegraph'
import { ThemeProps } from '../../../../shared/src/theme'
import { FileDecorator } from '../../tree/FileDecorator'

/**
 * Use a multi-column layout for tree entries when there are at least this many. See TreeEntriesSection.scss
 * for more information.
 */
const MIN_ENTRIES_FOR_COLUMN_LAYOUT = 6

const TreeEntry: React.FunctionComponent<{
    isDirectory: boolean
    name: string
    parentPath: string
    url: string
    isColumnLayout: boolean
    renderedFileDecorations: React.ReactNode
    path: string
}> = ({ isDirectory, name, url, isColumnLayout, renderedFileDecorations, path }) => (
    // TODO(tj): Limit width when not column layout
    // {
    //     'w-25': !isColumnLayout,
    // }
    <Link
        to={url}
        className={classNames(
            'tree-entry test-page-file-decorable',
            isDirectory && 'font-weight-bold',
            `test-tree-entry-${isDirectory ? 'directory' : 'file'}`
        )}
        title={path}
    >
        <div
            className={classNames(
                'd-flex align-items-center justify-content-between test-file-decorable-name overflow-hidden'
            )}
        >
            <span>
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

export const TreeEntriesSection: React.FunctionComponent<TreeEntriesSectionProps> = ({
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
    const renderedDecorationsByIndex = directChildren.map(entry =>
        FileDecorator({
            // If component is not specified, or it is 'page', render it.
            fileDecorations: fileDecorationsByPath[entry.path]?.filter(
                decoration => decoration?.component !== 'sidebar'
            ),
            isLightTheme,
        })
    )
    // If no ReactNode is truthy, we want to hide column-rule
    const noDecorations = !renderedDecorationsByIndex.some(decoration => !!decoration)

    const isColumnLayout = directChildren.length > MIN_ENTRIES_FOR_COLUMN_LAYOUT

    return (
        <div
            className={
                isColumnLayout
                    ? classNames('tree-entries-section--columns pr-2', {
                          'tree-entries-section--no-decorations': noDecorations,
                      })
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
