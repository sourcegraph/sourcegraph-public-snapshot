import * as React from 'react'

import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { FileDecoration } from 'sourcegraph'

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
import { FileDecorator } from './FileDecorator'
import { TreeLayerProps } from './TreeLayer'
import { treePadding } from './util'

interface TreeChildProps extends TreeLayerProps {
    className?: string
    maxEntries: number
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
export const Directory: React.FunctionComponent<React.PropsWithChildren<TreeChildProps>> = (
    props: TreeChildProps
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
                            <Icon as={props.isExpanded ? ChevronDownIcon : ChevronRightIcon} />
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
            {props.index === props.maxEntries - 1 && (
                <TreeRowAlert
                    variant="warning"
                    style={treePadding(props.depth, true, true)}
                    error="Too many entries. Use search to find a specific file."
                />
            )}
        </TreeLayerCell>
    </TreeRow>
)
