import React, { useCallback, useEffect } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import { flushSync } from 'react-dom'
import { useLocation, useNavigate } from 'react-router-dom'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { SymbolKind } from '@sourcegraph/shared/src/symbols/SymbolKind'
import { Link, flattenTree, type TreeNode } from '@sourcegraph/wildcard'

import { type SymbolNodeFields, SymbolKind as SymbolKindEnum } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

import { FocusableTree, type FocusableTreeProps } from './RepoRevisionSidebarFocusableTree'
import type { SymbolPlaceholder, SymbolWithChildren } from './RepoRevisionSidebarSymbols'

import styles from './RepoRevisionSidebarSymbols.module.scss'

interface Props extends FocusableTreeProps {
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
    focusKey,
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

            // We use flushSync here to make sure both the selectedSymbolUrl and the React router
            // navigation are applied at the same time. This fixed a race condition where one value
            // is updated before the other that caused the symbol tree to try and recover, jumping
            // between the two options indefinitely.
            //
            // Note: React errors when calling flushSync from lifecycle method, so we move this into
            // an async timeout instead.
            setTimeout(
                () =>
                    flushSync(() => {
                        onClick()
                        setSelectedSymbolUrl(element.url)
                        navigate(element.url)
                    }),
                0
            )
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
        <FocusableTree<SymbolNode>
            data={treeData}
            focusKey={focusKey}
            defaultExpandedIds={defaultExpandedIds}
            onSelect={onSelect}
            selectedIds={selectedIds}
            nodeClassName={styles.treeNode}
            renderNode={({
                element,
                handleSelect,
                props,
            }: {
                element: SymbolNode
                handleSelect: (event: React.MouseEvent) => {}
                props: { className: string; tabIndex: number }
            }): React.ReactNode => {
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
                        <span data-testid="symbol-name">{element.name}</span>
                    </Link>
                )
            }}
        />
    )
}
