/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-static-element-interactions: warn */
import * as React from 'react'
import { useEffect, useState } from 'react'

import { mdiSourceRepository, mdiFileDocumentOutline } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import { escapeRegExp, isEqual } from 'lodash'
import { NavLink } from 'react-router-dom'
import { FileDecoration } from 'sourcegraph'

import { gql, useQuery } from '@sourcegraph/http-client'
import { PrefetchableFile } from '@sourcegraph/shared/src/components/PrefetchableFile'
import { SymbolTag } from '@sourcegraph/shared/src/symbols/SymbolTag'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { InlineSymbolsResult } from '../graphql-operations'
import { fetchBlob, usePrefetchBlobFormat } from '../repo/blob/backend'
import { useExperimentalFeatures } from '../stores'
import { parseBrowserRepoURL } from '../util/url'

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
import { FileDecorator } from './FileDecorator'
import { useTreeRootContext } from './TreeContext'
import { TreeLayerProps } from './TreeLayer'
import { TreeEntryInfo, getTreeItemOffset } from './util'

import styles from './File.module.scss'
import treeStyles from './Tree.module.scss'

interface FileProps extends ThemeProps {
    fileDecorations?: FileDecoration[]
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
    enableMergedFileSymbolSidebar: boolean
    isGoUpTreeLink?: boolean

    // For core workflow inline symbols redesign
    location: H.Location
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
        fileDecorations,
        isLightTheme,
        depth,
        index,
        location,
        enableMergedFileSymbolSidebar,
        customIconPath,
    } = props

    const { revision, repoName } = useTreeRootContext()
    const isSidebarFilePrefetchEnabled = useExperimentalFeatures(
        features => features.enableSidebarFilePrefetch ?? false
    )
    const prefetchBlobFormat = usePrefetchBlobFormat()

    const renderedFileDecorations = (
        <FileDecorator
            // If component is not specified, or it is 'sidebar', render it.
            fileDecorations={fileDecorations?.filter(decoration => decoration?.where !== 'page')}
            isLightTheme={isLightTheme}
            isActive={isActive}
        />
    )

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
                                    {renderedFileDecorations}
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
                                    {renderedFileDecorations}
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
                                {renderedFileDecorations}
                            </TreeLayerRowContentsText>
                        </PrefetchableFile>
                    )}
                    {index === MAX_TREE_ENTRIES - 1 && (
                        <TreeRowAlert
                            variant="warning"
                            style={getTreeItemOffset(depth + 1)}
                            error="Too many entries. Use search to find a specific file."
                        />
                    )}
                </TreeLayerCell>
            </TreeRow>
            {enableMergedFileSymbolSidebar && isActive && (
                <Symbols activePath={entryInfo.path} location={location} style={offsetStyle} />
            )}
        </>
    )
}

export const SYMBOLS_QUERY = gql`
    query InlineSymbols($repo: ID!, $revision: String!, $includePatterns: [String!]) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revision) {
                    symbols(first: 1000, query: "", includePatterns: $includePatterns) {
                        ...InlineSymbolConnectionFields
                    }
                }
            }
        }
    }

    fragment InlineSymbolConnectionFields on SymbolConnection {
        __typename
        nodes {
            ...InlineSymbolNodeFields
        }
    }

    fragment InlineSymbolNodeFields on Symbol {
        __typename
        name
        containerName
        kind
        language
        location {
            resource {
                path
            }
            range {
                start {
                    line
                    character
                }
                end {
                    line
                    character
                }
            }
        }
        url
    }
`

interface SymbolsProps
    extends Pick<TreeLayerProps, 'activePath' | 'location'>,
        Pick<React.HTMLAttributes<HTMLDivElement>, 'style'> {}

const Symbols: React.FunctionComponent<SymbolsProps> = ({ activePath, location, style }) => {
    const { repoID, revision } = useTreeRootContext()
    const { data, loading, error } = useQuery<InlineSymbolsResult>(SYMBOLS_QUERY, {
        variables: {
            repo: repoID,
            revision,
            // `includePatterns` expects regexes, so first escape the path.
            includePatterns: ['^' + escapeRegExp(activePath)],
        },
    })

    const currentLocation = parseBrowserRepoURL(H.createPath(location))
    const isSymbolActive = (symbolUrl: string): boolean => {
        const symbolLocation = parseBrowserRepoURL(symbolUrl)
        return (
            currentLocation.repoName === symbolLocation.repoName &&
            currentLocation.revision === symbolLocation.revision &&
            currentLocation.filePath === symbolLocation.filePath &&
            isEqual(currentLocation.position, symbolLocation.position)
        )
    }

    if (loading) {
        return (
            <Delay timeout={800}>
                <TreeRow className={styles.symbols}>
                    <TreeLayerCell>
                        <TreeLayerRowContents className="d-flex" style={style}>
                            <LoadingSpinner /> Loading symbol data...
                        </TreeLayerRowContents>
                    </TreeLayerCell>
                </TreeRow>
            </Delay>
        )
    }

    let content = null

    if (error) {
        content = 'Unable to load symbol data.'
    }

    if (data && data.node?.__typename === 'Repository') {
        // Only consider top-level symbols
        const symbols = data.node.commit?.symbols.nodes.filter(symbol => !symbol.containerName) ?? []
        if (symbols.length > 0) {
            content = (
                <ul>
                    {symbols.map(symbol => (
                        <li key={symbol.url}>
                            <NavLink
                                to={symbol.url}
                                isActive={() => isSymbolActive(symbol.url)}
                                className={classNames('test-symbol-link', styles.link)}
                                activeClassName={styles.linkActive}
                            >
                                <SymbolTag kind={symbol.kind} className="mr-1 test-symbol-icon" />
                                <span className={classNames('test-symbol-name')}>{symbol.name}</span>
                            </NavLink>
                        </li>
                    ))}
                </ul>
            )
        } else {
            content = 'No symbols found.'
        }
    }

    if (content) {
        return (
            <TreeRow className={styles.symbols}>
                <TreeLayerCell>
                    <TreeLayerRowContents style={style}>{content}</TreeLayerRowContents>
                </TreeLayerCell>
            </TreeRow>
        )
    }
    return null
}

const Delay: React.FunctionComponent<{ timeout: number; children: React.ReactElement }> = ({ timeout, children }) => {
    const [show, setShow] = useState(false)

    useEffect(() => {
        const id = setTimeout(() => setShow(true), timeout)
        return () => clearTimeout(id)
    }, [timeout])

    return show ? children : null
}
