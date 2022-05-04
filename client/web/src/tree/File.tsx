/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-static-element-interactions: warn */
import * as React from 'react'

import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { Icon } from '@sourcegraph/wildcard'

import {
    TreeLayerCell,
    TreeLayerRowContents,
    TreeLayerRowContentsLink,
    TreeRowAlert,
    TreeLayerRowContentsText,
    TreeRowIcon,
    TreeRowLabel,
    TreeRow,
} from './components'
import { FileDecorator } from './FileDecorator'
import { TreeLayerProps } from './TreeLayer'
import { maxEntries, treePadding } from './util'

interface FileProps extends TreeLayerProps {
    className?: string
    maxEntries: number
    handleTreeClick: () => void
    noopRowClick: (event: React.MouseEvent<HTMLAnchorElement>) => void
    linkRowClick: (event: React.MouseEvent<HTMLAnchorElement>) => void
    isActive: boolean
    isSelected: boolean
}

export const File: React.FunctionComponent<React.PropsWithChildren<FileProps>> = props => {
    const renderedFileDecorations = (
        <FileDecorator
            // If component is not specified, or it is 'sidebar', render it.
            fileDecorations={props.fileDecorations?.filter(decoration => decoration?.where !== 'page')}
            isLightTheme={props.isLightTheme}
            isActive={props.isActive}
        />
    )

    return (
        <TreeRow
            key={props.entryInfo.path}
            className={props.className}
            isActive={props.isActive}
            isSelected={props.isSelected}
        >
            <TreeLayerCell className="test-sidebar-file-decorable">
                {props.entryInfo.submodule ? (
                    props.entryInfo.url ? (
                        <TreeLayerRowContentsLink
                            to={props.entryInfo.url}
                            onClick={props.linkRowClick}
                            draggable={false}
                            title={'Submodule: ' + props.entryInfo.submodule.url}
                            data-tree-path={props.entryInfo.path}
                        >
                            <TreeLayerRowContentsText>
                                {/* TODO Improve accessibility: https://github.com/sourcegraph/sourcegraph/issues/12916 */}
                                <TreeRowIcon style={treePadding(props.depth, true)} onClick={props.noopRowClick}>
                                    <Icon as={SourceRepositoryIcon} />
                                </TreeRowIcon>
                                <TreeRowLabel className="test-file-decorable-name">
                                    {props.entryInfo.name} @ {props.entryInfo.submodule.commit.slice(0, 7)}
                                </TreeRowLabel>
                                {renderedFileDecorations}
                            </TreeLayerRowContentsText>
                        </TreeLayerRowContentsLink>
                    ) : (
                        <TreeLayerRowContents title={'Submodule: ' + props.entryInfo.submodule.url}>
                            <TreeLayerRowContentsText>
                                <TreeRowIcon style={treePadding(props.depth, true)}>
                                    <Icon as={SourceRepositoryIcon} />
                                </TreeRowIcon>
                                <TreeRowLabel className="test-file-decorable-name">
                                    {props.entryInfo.name} @ {props.entryInfo.submodule.commit.slice(0, 7)}
                                </TreeRowLabel>
                                {renderedFileDecorations}
                            </TreeLayerRowContentsText>
                        </TreeLayerRowContents>
                    )
                ) : (
                    <TreeLayerRowContentsLink
                        className="test-tree-file-link"
                        to={props.entryInfo.url}
                        onClick={props.linkRowClick}
                        data-tree-path={props.entryInfo.path}
                        draggable={false}
                        title={props.entryInfo.path}
                        // needed because of dynamic styling
                        style={treePadding(props.depth, false)}
                        tabIndex={-1}
                    >
                        <TreeLayerRowContentsText className="d-flex flex-row flex-1 justify-content-between">
                            <span className="test-file-decorable-name">{props.entryInfo.name}</span>
                            {renderedFileDecorations}
                        </TreeLayerRowContentsText>
                    </TreeLayerRowContentsLink>
                )}
                {props.index === maxEntries - 1 && (
                    <TreeRowAlert
                        variant="warning"
                        style={treePadding(props.depth, true)}
                        error="Too many entries. Use search to find a specific file."
                    />
                )}
            </TreeLayerCell>
        </TreeRow>
    )
}
