import * as React from 'react'

import { mdiChevronDown, mdiChevronRight } from '@mdi/js'
import { FileDecoration } from 'sourcegraph'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner, Icon } from '@sourcegraph/wildcard'

import {
    TreeLayerRowContentsText,
    TreeLayerCell,
    TreeRowAlert,
    TreeLayerRowContents,
    TreeRowIconLink,
    TreeRowLabelLink,
    TreeRow,
} from './components'
import { MAX_TREE_ENTRIES } from './constants'
import { FileDecorator } from './FileDecorator'
import { TreeEntryInfo, treePadding } from './util'

interface DirectoryProps extends ThemeProps {
    depth: number
    index: number
    className?: string
    entryInfo: TreeEntryInfo
    isExpanded: boolean
    loading: boolean
    handleTreeClick: () => void
    noopRowClick: (event: React.MouseEvent<HTMLAnchorElement>) => void
    linkRowClick: (event: React.MouseEvent<HTMLAnchorElement>) => void
    fileDecorations?: FileDecoration[]
    isActive: boolean
    isSelected: boolean
}

/**
 * JSX to render a tree directory
 *
 * @param props
 */
export const Directory: React.FunctionComponent<React.PropsWithChildren<DirectoryProps>> = (
    props: DirectoryProps
): JSX.Element => (
    <TreeRow
        key={props.entryInfo.path}
        className={props.className}
        onClick={props.handleTreeClick}
        isActive={props.isActive}
        isSelected={props.isSelected}
        isExpanded={props.isExpanded}
    >
        <TreeLayerCell className="test-sidebar-file-decorable">
            <TreeLayerRowContents data-tree-is-directory="true" data-tree-path={props.entryInfo.path} isNew={true}>
                <TreeLayerRowContentsText className="flex-1 justify-between">
                    <div className="d-flex">
                        <TreeRowIconLink
                            style={treePadding(props.depth, true, true)}
                            className="test-tree-noop-link"
                            href={props.entryInfo.url}
                            onClick={props.noopRowClick}
                            tabIndex={-1}
                        >
                            <Icon svgPath={props.isExpanded ? mdiChevronDown : mdiChevronRight} aria-hidden={true} />
                        </TreeRowIconLink>
                        <TreeRowLabelLink
                            to={props.entryInfo.url}
                            onClick={props.linkRowClick}
                            className="test-file-decorable-name"
                            draggable={false}
                            title={props.entryInfo.path}
                            tabIndex={-1}
                        >
                            {props.entryInfo.name}
                        </TreeRowLabelLink>
                    </div>
                    <FileDecorator
                        // If component is not specified, or it is 'sidebar', render it.
                        fileDecorations={props.fileDecorations?.filter(decoration => decoration?.where !== 'page')}
                        className="mr-3"
                        isLightTheme={props.isLightTheme}
                        isActive={props.isActive}
                    />
                </TreeLayerRowContentsText>
                {props.loading && (
                    <div>
                        <LoadingSpinner className="tree-page__entries-loader" />
                    </div>
                )}
            </TreeLayerRowContents>
            {props.index === MAX_TREE_ENTRIES - 1 && (
                <TreeRowAlert
                    variant="warning"
                    style={treePadding(props.depth, true, true)}
                    error="Too many entries. Use search to find a specific file."
                />
            )}
        </TreeLayerCell>
    </TreeRow>
)
