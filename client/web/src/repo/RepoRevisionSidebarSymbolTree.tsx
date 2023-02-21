import React, { useCallback, useEffect } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import { useLocation, useNavigate } from 'react-router-dom'

import { SymbolKind } from '@sourcegraph/shared/src/symbols/SymbolKind'
import { Tree, Link, flattenTree, TreeNode } from '@sourcegraph/wildcard'

import { SymbolNodeFields, SymbolKind as SymbolKindEnum } from '../graphql-operations'
import { useExperimentalFeatures } from '../stores'
import { parseBrowserRepoURL } from '../util/url'

import type { SymbolPlaceholder, SymbolWithChildren } from './RepoRevisionSidebarSymbols'

interface Props {
    symbols: SymbolWithChildren[]
    onClick: () => void
    selectedSymbolUrl: string | null
    setSelectedSymbolUrl: (url: string | null) => void
}
type SymbolNode = (Omit<SymbolNodeFields, 'children'> & TreeNode) | (Omit<SymbolPlaceholder, 'children'> & TreeNode)

export const RepoRevisionSidebarSymbolTree: React.FC<Props> = ({
    symbols,
    onClick,
    selectedSymbolUrl,
    setSelectedSymbolUrl,
}) => {
    const symbolKindTags = useExperimentalFeatures(features => features.symbolKindTags)

    // @ts-expect-error: The intersection and Omit combination breaks the union
    const treeData: SymbolNode[] = flattenTree<SymbolWithChildren>({
        __typename: 'SymbolPlaceholder',
        name: 'root',
        children: symbols,
    })

    const location = useLocation()
    const navigate = useNavigate()

    // Mark all expandable nodes as expanded by default.
    const defaultExpandedIds = treeData.reduce((acc, node) => {
        if (node.children.length > 0) {
            acc.push(node.id)
        }
        return acc
    }, [] as number[])

    const onSelect = useCallback(
        ({ element, isSelected }: { element: SymbolNode; isSelected: boolean }) => {
            if (!isSelected || element.__typename !== 'Symbol' || selectedSymbolUrl === element.url) {
                return
            }

            navigate(element.url)
            onClick()
            setSelectedSymbolUrl(element.url)
        },
        [navigate, onClick, selectedSymbolUrl, setSelectedSymbolUrl]
    )

    // We need a mutable reference to the tree data since we don't want some
    // hooks to run when the tree data changes.
    const treeDataRef = React.useRef(treeData)
    useEffect(() => {
        treeDataRef.current = treeData
    }, [treeData])
    useEffect(() => {
        const currentLocation = parseBrowserRepoURL(H.createPath(location))

        // Ignore effect updates that are caused by changes that we have made
        if (selectedSymbolUrl && isEqual(currentLocation, parseBrowserRepoURL(selectedSymbolUrl))) {
            return
        }

        function isSymbolActive(symbolUrl: string): boolean {
            const symbolLocation = parseBrowserRepoURL(symbolUrl)
            return (
                currentLocation.repoName === symbolLocation.repoName &&
                currentLocation.revision === symbolLocation.revision &&
                currentLocation.filePath === symbolLocation.filePath &&
                isEqual(currentLocation.position, symbolLocation.position)
            )
        }

        for (const node of treeDataRef.current) {
            if (node.__typename === 'SymbolPlaceholder') {
                continue
            }
            if (isSymbolActive(node.url)) {
                if (selectedSymbolUrl !== node.url) {
                    setSelectedSymbolUrl(node.url)
                }
                break
            }
        }
    }, [location, selectedSymbolUrl, setSelectedSymbolUrl])

    const selectedIds = selectedSymbolUrl
        ? treeData.filter(node => node.__typename === 'Symbol' && node.url === selectedSymbolUrl).map(node => node.id)
        : []

    return (
        <Tree<SymbolNode>
            data={treeData}
            defaultExpandedIds={defaultExpandedIds}
            onSelect={onSelect}
            selectedIds={selectedIds}
            renderNode={({ element, handleSelect, props }): React.ReactNode => {
                const { className, ...rest } = props

                const to = element.__typename === 'SymbolPlaceholder' ? '' : element.url
                const kind = element.__typename === 'SymbolPlaceholder' ? SymbolKindEnum.UNKNOWN : element.kind

                return (
                    <Link
                        {...rest}
                        className={classNames(className, 'test-symbol-link')}
                        to={to}
                        onClick={event => {
                            event.preventDefault()
                            handleSelect(event)
                        }}
                    >
                        <SymbolKind kind={kind} className="mr-1" symbolKindTags={symbolKindTags} />
                        {element.name}
                    </Link>
                )
            }}
        />
    )
}
