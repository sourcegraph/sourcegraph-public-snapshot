import React from 'react'

import { mdiFolderOutline } from '@mdi/js'

import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { dirname } from '../util/path'

import { TreeLayerTable } from './components'
import { GO_UP_TREE_LABEL } from './constants'
import { File } from './File'
import { SingleChildTreeLayer } from './SingleChildTreeLayer'
import { TreeRootContext } from './TreeContext'
import { TreeLayer } from './TreeLayer'
import { TreeRootProps } from './TreeRoot'
import { hasSingleChild, NOOP, SingleChildGitTree, TreeEntryInfo } from './util'

interface ChildTreeLayerProps extends Pick<TreeRootProps, Exclude<keyof TreeRootProps, 'sizeKey'>> {
    fileDecorationsByPath: FileDecorationsByPath

    entries: TreeEntryInfo[]
    /** The entry information for a SingleChildTreeLayer. Will be a SingleChildGitTree with fields populated, or be an empty object if there is no SingleChildTreeLayer to render. */
    singleChildTreeEntry: SingleChildGitTree
    /** The children entries of a SingleChildTreeLayer. Will be undefined if there is no SingleChildTreeLayer to render. */
    childrenEntries?: SingleChildGitTree[]
    enableMergedFileSymbolSidebar: boolean
    onHover: (filePath: string) => void
}

/**
 * Either a SingleChildTreeLayer or TreeLayer.
 */
export const ChildTreeLayer: React.FunctionComponent<React.PropsWithChildren<ChildTreeLayerProps>> = (
    props: ChildTreeLayerProps
) => {
    const sharedProps = {
        location: props.location,
        activePath: props.activePath,
        activeNode: props.activeNode,
        depth: props.depth + 1,
        expandedTrees: props.expandedTrees,
        parent: props.parent,
        repoName: props.repoName,
        repoID: props.repoID,
        revision: props.revision,
        onToggleExpand: props.onToggleExpand,
        onHover: props.onHover,
        selectedNode: props.selectedNode,
        setChildNodes: props.setChildNodes,
        setActiveNode: props.setActiveNode,
        onSelect: props.onSelect,
        commitID: props.commitID,
        extensionsController: props.extensionsController,
        isLightTheme: props.isLightTheme,
    }

    // Only show ".." (go up) for non-root file trees
    const shouldShowGoUp = props.depth === -1 && props.parentPath

    return (
        <div>
            <TreeLayerTable>
                <tbody>
                    {shouldShowGoUp && (
                        <tr>
                            <td>
                                <TreeLayerTable>
                                    <tbody>
                                        <TreeRootContext.Consumer>
                                            {treeRootContext => (
                                                <File
                                                    entryInfo={{
                                                        name: GO_UP_TREE_LABEL,
                                                        path: props.parentPath as string,
                                                        isDirectory: false,
                                                        url: dirname(treeRootContext.rootTreeUrl),
                                                        isSingleChild: false,
                                                        submodule: null,
                                                    }}
                                                    location={props.location}
                                                    depth={sharedProps.depth}
                                                    index={0}
                                                    isLightTheme={sharedProps.isLightTheme}
                                                    handleTreeClick={NOOP}
                                                    noopRowClick={NOOP}
                                                    linkRowClick={() => props.telemetryService.log('FileTreeClick')}
                                                    isActive={false}
                                                    isSelected={false}
                                                    isGoUpTreeLink={true}
                                                    customIconPath={mdiFolderOutline}
                                                    enableMergedFileSymbolSidebar={props.enableMergedFileSymbolSidebar}
                                                />
                                            )}
                                        </TreeRootContext.Consumer>
                                    </tbody>
                                </TreeLayerTable>
                            </td>
                        </tr>
                    )}
                    <tr>
                        <td>
                            {hasSingleChild(props.entries) ? (
                                <SingleChildTreeLayer
                                    {...sharedProps}
                                    key={props.singleChildTreeEntry.path}
                                    index={shouldShowGoUp ? 1 : 0}
                                    isExpanded={props.expandedTrees.includes(props.singleChildTreeEntry.path)}
                                    parentPath={props.singleChildTreeEntry.path}
                                    entryInfo={props.singleChildTreeEntry}
                                    childrenEntries={props.singleChildTreeEntry.children}
                                    fileDecorationsByPath={props.fileDecorationsByPath}
                                    fileDecorations={props.fileDecorationsByPath[props.singleChildTreeEntry.path]}
                                    telemetryService={props.telemetryService}
                                    enableMergedFileSymbolSidebar={props.enableMergedFileSymbolSidebar}
                                />
                            ) : (
                                props.entries.map((item, index) => (
                                    <TreeLayer
                                        {...sharedProps}
                                        key={item.path}
                                        index={shouldShowGoUp ? index + 1 : index}
                                        isExpanded={props.expandedTrees.includes(item.path)}
                                        parentPath={item.path}
                                        entryInfo={item}
                                        fileDecorations={props.fileDecorationsByPath[item.path]}
                                        telemetryService={props.telemetryService}
                                        enableMergedFileSymbolSidebar={props.enableMergedFileSymbolSidebar}
                                    />
                                ))
                            )}
                        </td>
                    </tr>
                </tbody>
            </TreeLayerTable>
        </div>
    )
}
