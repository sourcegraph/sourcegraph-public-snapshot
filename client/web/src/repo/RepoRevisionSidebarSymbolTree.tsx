import React, { useCallback, useEffect } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import { flushSync } from 'react-dom'
import { useLocation, useNavigate } from 'react-router-dom'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { SymbolKind } from '@sourcegraph/shared/src/symbols/SymbolKind'
import { ParsedRepoRevision, ParsedRepoURI, UIPositionSpec } from '@sourcegraph/shared/src/util/url'
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
        const currentParsedURL = parseBrowserRepoURL(H.createPath(location))
        // Ignore effect updates that are caused by changes that we have made.
        if (
            (selectedSymbolUrl && isEqual(currentParsedURL, parseBrowserRepoURL(selectedSymbolUrl))) ||
            !currentParsedURL.position
        ) {
            return
        }

        type Location = ParsedRepoURI & UIPositionSpec & Pick<ParsedRepoRevision, 'rawRevision'>
        const currentLocation: Location = { ...currentParsedURL, position: currentParsedURL.position }

        function isSameFile(l1: Location, l2: Location): boolean {
            return l1.repoName === l2.repoName && l1.filePath === l2.filePath && l1.revision === l2.revision
        }

        function isSymbolActive(symbolLocation: Location): boolean {
            return currentLocation.position.line === symbolLocation.position.line
        }

        // Sometimes current location will not match the exact position of a known symbol,
        // but will be inside some higher-level container instead (function body, struct, etc.),
        // in which case we can highlight the nearest symbol above.
        const nearest: { line: number; url?: string; location?: Location } = { line: -1 }
        function updateNearestSymbol(symbolLocation: Location, url: string): undefined {
            const { line } = symbolLocation.position
            const { line: anchor } = currentLocation.position!
            if (line < anchor && line > nearest.line) {
                nearest.line = line
                nearest.url = url
                nearest.location = symbolLocation
            }
        }

        for (const node of treeDataRef.current) {
            if (node.__typename === 'SymbolPlaceholder') {
                continue
            }
            const parsedURL = parseBrowserRepoURL(node.url)
            if (!parsedURL.position) {
                continue
            }
            const symbolLocation = { ...parsedURL, position: parsedURL.position }
            if (!isSameFile(currentLocation, symbolLocation)) {
                continue
            }
            if (isSymbolActive(symbolLocation)) {
                if (selectedSymbolUrl !== node.url) {
                    setSelectedSymbolUrl(node.url)
                }
                return
            }
            updateNearestSymbol(symbolLocation, node.url)
        }

        if (nearest.line >= 0 && nearest.url) {
            setSelectedSymbolUrl(nearest.url)
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
