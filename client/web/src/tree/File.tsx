/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-static-element-interactions: warn */
import * as React from 'react'

import { mdiSourceRepository, mdiFileDocumentOutline } from '@mdi/js'

import { PrefetchableFile } from '@sourcegraph/shared/src/components/PrefetchableFile'
import { Icon } from '@sourcegraph/wildcard'

import { fetchBlob, usePrefetchBlobFormat } from '../repo/blob/backend'
import { useExperimentalFeatures } from '../stores'

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
import { MAX_TREE_ENTRIES } from './constants'
import { useTreeRootContext } from './TreeContext'
import { TreeEntryInfo, getTreeItemOffset } from './util'

import treeStyles from './Tree.module.scss'

interface FileProps {
    entryInfo: TreeEntryInfo
    depth: number
    index: number
    className?: string
    handleTreeClick: () => void
    noopRowClick: (event: React.MouseEvent<HTMLAnchorElement>) => void
    linkRowClick: (event: React.MouseEvent<HTMLAnchorElement>) => void
    isActive: boolean
    isSelected: boolean
    customIconPath?: string
    isGoUpTreeLink?: boolean
}

export const File: React.FunctionComponent<React.PropsWithChildren<FileProps>> = props => {
    const {
        isActive,
        isSelected,
        isGoUpTreeLink,
        className,
        entryInfo,
        linkRowClick,
        noopRowClick,
        depth,
        index,
        customIconPath,
    } = props

    const { revision, repoName } = useTreeRootContext()
    const isSidebarFilePrefetchEnabled = useExperimentalFeatures(
        features => features.enableSidebarFilePrefetch ?? false
    )
    const prefetchBlobFormat = usePrefetchBlobFormat()

    const offsetStyle = getTreeItemOffset(depth)

    return (
        <>
            <TreeRow key={entryInfo.path} className={className} isActive={isActive} isSelected={isSelected}>
                <TreeLayerCell className="test-sidebar-file-decorable">
                    {entryInfo.submodule ? (
                        entryInfo.url ? (
                            <TreeLayerRowContentsLink
                                to={entryInfo.url}
                                onClick={linkRowClick}
                                draggable={false}
                                title={'Submodule: ' + entryInfo.submodule.url}
                                data-tree-path={entryInfo.path}
                            >
                                <TreeLayerRowContentsText>
                                    {/* TODO Improve accessibility: https://github.com/sourcegraph/sourcegraph/issues/12916 */}
                                    <TreeRowIcon style={offsetStyle} onClick={noopRowClick}>
                                        <Icon aria-hidden={true} svgPath={mdiSourceRepository} />
                                    </TreeRowIcon>
                                    <TreeRowLabel className="test-file-decorable-name">
                                        {entryInfo.name} @ {entryInfo.submodule.commit.slice(0, 7)}
                                    </TreeRowLabel>
                                </TreeLayerRowContentsText>
                            </TreeLayerRowContentsLink>
                        ) : (
                            <TreeLayerRowContents title={'Submodule: ' + entryInfo.submodule.url}>
                                <TreeLayerRowContentsText>
                                    <TreeRowIcon style={offsetStyle}>
                                        <Icon aria-hidden={true} svgPath={mdiSourceRepository} />
                                    </TreeRowIcon>
                                    <TreeRowLabel className="test-file-decorable-name">
                                        {entryInfo.name} @ {entryInfo.submodule.commit.slice(0, 7)}
                                    </TreeRowLabel>
                                </TreeLayerRowContentsText>
                            </TreeLayerRowContents>
                        )
                    ) : (
                        <PrefetchableFile
                            isPrefetchEnabled={isSidebarFilePrefetchEnabled && !isActive && !isGoUpTreeLink}
                            prefetch={params =>
                                fetchBlob({
                                    ...params,
                                    format: prefetchBlobFormat,
                                })
                            }
                            isSelected={isSelected}
                            revision={revision}
                            repoName={repoName}
                            filePath={entryInfo.path}
                            as={TreeLayerRowContentsLink}
                            className="test-tree-file-link"
                            to={entryInfo.url}
                            onClick={linkRowClick}
                            data-tree-path={entryInfo.path}
                            draggable={false}
                            title={entryInfo.path}
                            // needed because of dynamic styling
                            style={offsetStyle}
                            tabIndex={-1}
                        >
                            <TreeLayerRowContentsText className="d-flex">
                                <TreeRowIcon onClick={noopRowClick}>
                                    <Icon
                                        className={treeStyles.treeIcon}
                                        svgPath={customIconPath || mdiFileDocumentOutline}
                                        aria-hidden={true}
                                    />
                                </TreeRowIcon>
                                <TreeRowLabel className="test-file-decorable-name">{entryInfo.name}</TreeRowLabel>
                            </TreeLayerRowContentsText>
                        </PrefetchableFile>
                    )}
                    {index === MAX_TREE_ENTRIES - 1 && (
                        <TreeRowAlert
                            variant="note"
                            style={getTreeItemOffset(depth + 1)}
                            error="Full list of files is too long to display. Use search to find specific file."
                        />
                    )}
                </TreeLayerCell>
            </TreeRow>
        </>
    )
}
