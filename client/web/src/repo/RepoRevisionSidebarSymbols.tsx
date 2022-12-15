import React, { useState, useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { entries, escapeRegExp, flatMap, flow, groupBy, isEqual } from 'lodash/fp'
import { NavLink, useLocation } from 'react-router-dom'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { logger } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { SymbolKind as SymbolKindEnum } from '@sourcegraph/shared/src/schema'
import { SymbolKind } from '@sourcegraph/shared/src/symbols/SymbolKind'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Alert, useDebounce } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionForm,
    ConnectionContainer,
    ConnectionLoading,
    ConnectionSummary,
    SummaryContainer,
    ShowMoreButton,
} from '../components/FilteredConnection/ui'
import { Scalars, SymbolNodeFields, SymbolsResult, SymbolsVariables } from '../graphql-operations'
import { useExperimentalFeatures } from '../stores'
import { parseBrowserRepoURL } from '../util/url'

import styles from './RepoRevisionSidebarSymbols.module.scss'

interface SymbolNodeProps {
    node: SymbolWithChildren
    onHandleClick: () => void
    isActive: boolean
    nestedRender: HierarchicalSymbolsProps['render']
}

const SymbolNode: React.FunctionComponent<React.PropsWithChildren<SymbolNodeProps>> = ({
    node,
    onHandleClick,
    isActive,
    nestedRender,
}) => {
    const isActiveFunc = (): boolean => isActive
    const symbolKindTags = useExperimentalFeatures(features => features.symbolKindTags)
    return (
        <li className={styles.repoRevisionSidebarSymbolsNode}>
            {node.__typename === 'SymbolPlaceholder' ? (
                <span className={styles.link}>
                    <SymbolKind kind={SymbolKindEnum.UNKNOWN} className="mr-1" symbolKindTags={symbolKindTags} />
                    {node.name}
                </span>
            ) : (
                <NavLink
                    to={node.url}
                    isActive={isActiveFunc}
                    className={classNames('test-symbol-link', styles.link)}
                    activeClassName={styles.linkActive}
                    onClick={onHandleClick}
                >
                    <SymbolKind kind={node.kind} className="mr-1" symbolKindTags={symbolKindTags} />
                    <span className={styles.name} data-testid="symbol-name">
                        {node.name}
                    </span>
                </NavLink>
            )}
            {node.children && <HierarchicalSymbols symbols={node.children} render={nestedRender} className="pl-3" />}
        </li>
    )
}

export const SYMBOLS_QUERY = gql`
    query Symbols($repo: ID!, $revision: String!, $first: Int, $query: String, $includePatterns: [String!]) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revision) {
                    symbols(first: $first, query: $query, includePatterns: $includePatterns) {
                        ...SymbolConnectionFields
                    }
                }
            }
        }
    }

    fragment SymbolConnectionFields on SymbolConnection {
        __typename
        pageInfo {
            hasNextPage
        }
        nodes {
            ...SymbolNodeFields
        }
    }

    fragment SymbolNodeFields on Symbol {
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

const BATCH_COUNT = 100

export interface RepoRevisionSidebarSymbolsProps extends Partial<RevisionSpec> {
    repoID: Scalars['ID']
    /** The path of the file or directory currently shown in the content area */
    activePath: string
    onHandleSymbolClick: () => void
}

export const RepoRevisionSidebarSymbols: React.FunctionComponent<
    React.PropsWithChildren<RepoRevisionSidebarSymbolsProps>
> = ({ repoID, revision = '', activePath, onHandleSymbolClick }) => {
    const location = useLocation()
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)

    const { connection, error, loading, hasNextPage, fetchMore } = useShowMorePagination<
        SymbolsResult,
        SymbolsVariables,
        SymbolNodeFields
    >({
        query: SYMBOLS_QUERY,
        variables: {
            query,
            first: BATCH_COUNT,
            repo: repoID,
            revision,
            // `includePatterns` expects regexes, so first escape the path.
            includePatterns: ['^' + escapeRegExp(activePath)],
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                return { nodes: [] }
            }
            if (node.__typename !== 'Repository') {
                return { nodes: [] }
            }
            if (!node.commit?.symbols?.nodes) {
                return { nodes: [] }
            }

            return node.commit.symbols
        },
        options: {
            fetchPolicy: 'cache-first',
        },
    })

    const summary = connection && (
        <ConnectionSummary
            connection={connection}
            first={BATCH_COUNT}
            noun="symbol"
            pluralNoun="symbols"
            hasNextPage={hasNextPage}
            connectionQuery={query}
            compact={true}
        />
    )

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

    const heirarchicalSymbols = useMemo(
        () =>
            flow(
                groupBy<SymbolNodeFields>(symbol => symbol.location.resource.path),
                entries,
                flatMap(([, symbols]) => hierarchyOf(symbols))
            )(connection?.nodes ?? []),
        [connection?.nodes]
    )

    return (
        <ConnectionContainer className={classNames('h-100', styles.repoRevisionSidebarSymbols)} compact={true}>
            <div className={styles.formContainer}>
                <ConnectionForm
                    inputValue={searchValue}
                    onInputChange={event => setSearchValue(event.target.value)}
                    inputPlaceholder="Search symbols..."
                    compact={true}
                    formClassName={styles.form}
                />
                <SummaryContainer compact={true} className={styles.summaryContainer}>
                    {query && summary}
                </SummaryContainer>
            </div>
            {error && (
                <Alert variant={error.message.includes('Estimated completion') ? 'info' : 'danger'}>
                    <ErrorMessage error={error.message} />
                </Alert>
            )}
            {connection && (
                <HierarchicalSymbols
                    symbols={heirarchicalSymbols}
                    render={args => (
                        <SymbolNode
                            node={args.symbol}
                            onHandleClick={onHandleSymbolClick}
                            isActive={args.symbol.__typename === 'Symbol' && isSymbolActive(args.symbol.url)}
                            nestedRender={args.nestedRender}
                        />
                    )}
                />
            )}
            {loading && <ConnectionLoading compact={true} />}
            {!loading && connection && (
                <SummaryContainer compact={true} className={styles.summaryContainer}>
                    {!query && summary}
                    {hasNextPage && <ShowMoreButton compact={true} onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}

interface HierarchicalSymbolsProps {
    symbols: SymbolWithChildren[]
    render: (props: {
        symbol: SymbolWithChildren
        nestedRender: HierarchicalSymbolsProps['render']
    }) => React.ReactElement
    className?: string
}

const HierarchicalSymbols: React.FunctionComponent<HierarchicalSymbolsProps> = props => (
    <ul className={classNames(styles.hierarchicalSymbolsContainer, props.className)}>
        {props.symbols.map(symbol => (
            <props.render
                key={'url' in symbol ? symbol.url : fullName(symbol)}
                symbol={symbol}
                nestedRender={props.render}
            />
        ))}
    </ul>
)

// When searching symbols, results may contain only child symbols without their parents
// (e.g. when searching for "bar", a class named "Foo" with a method named "bar" will
// return "bar" as a result, and "bar" will say that "Foo" is its parent).
// The placeholder symbols exist to show the hierarchy of the results, but these placeholders
// are not interactive (cannot be clicked to navigate) and don't have any other information.
interface SymbolPlaceholder {
    __typename: 'SymbolPlaceholder'
    name: string
}

type SymbolWithChildren = (SymbolNodeFields | SymbolPlaceholder) & { children: SymbolWithChildren[] }

const hierarchyOf = (symbols: SymbolNodeFields[]): SymbolWithChildren[] => {
    const fullNameToSymbol = new Map<string, SymbolNodeFields>(symbols.map(symbol => [fullName(symbol), symbol]))
    const fullNameToSymbolWithChildren = new Map<string, SymbolWithChildren>()
    const topLevelSymbols: SymbolWithChildren[] = []

    const visit = (fullName: string): void => {
        if (fullName === '') {
            return
        }

        let symbol: SymbolNodeFields | SymbolPlaceholder | undefined = fullNameToSymbol.get(fullName)
        if (!symbol) {
            // Symbol doesn't exist, create placeholder at the top level and add current symbol to it.
            // (This happens when running a search and the result is a child of a symbol that isn't in the result set.)
            symbol = {
                __typename: 'SymbolPlaceholder',
                name: fullName.split('.').at(-1) || fullName,
            }
        }

        // symbolWithChildren might already exist if we've already visited a child of this symbol.
        const symbolWithChildren = fullNameToSymbolWithChildren.get(fullName) || { ...symbol, children: [] }
        fullNameToSymbolWithChildren.set(fullName, symbolWithChildren)

        const parentFullName =
            symbol.__typename === 'Symbol' ? symbol.containerName : fullName.split('.').slice(0, -1).join('.')
        if (!parentFullName) {
            // No parent, add to top-level
            topLevelSymbols.push(symbolWithChildren)
            return
        }

        const parentSymbol = fullNameToSymbol.get(parentFullName)
        let parentSymbolWithChildren = fullNameToSymbolWithChildren.get(parentFullName)
        if (parentSymbolWithChildren) {
            // Parent exists, add to parent
            parentSymbolWithChildren.children.push(symbolWithChildren)
        } else if (parentSymbol) {
            // Create parent node and add current symbol to it
            fullNameToSymbolWithChildren.set(parentFullName, { ...parentSymbol, children: [symbolWithChildren] })
        } else {
            // Parent doesn't exist, visit it to generate a placeholder hierarchy, then add current symbol to it
            visit(parentFullName)
            parentSymbolWithChildren = fullNameToSymbolWithChildren.get(parentFullName) // This should now exist
            if (parentSymbolWithChildren) {
                parentSymbolWithChildren.children.push(symbolWithChildren)
            } else {
                // This should never happen!!
                logger.error('RepoRevisionSidebarSymbols: Failed to add symbol to parent', { fullName, parentFullName })
            }
        }
    }

    for (const symbol of symbols) {
        visit(fullName(symbol))
    }

    // Sort everything
    for (const sym of fullNameToSymbolWithChildren.values()) {
        sym.children.sort((a, b) => a.name.localeCompare(b.name))
    }
    topLevelSymbols.sort((a, b) => a.name.localeCompare(b.name))

    return topLevelSymbols
}

const fullName = (symbol: SymbolNodeFields | SymbolPlaceholder): string =>
    `${symbol.__typename === 'Symbol' && symbol.containerName ? symbol.containerName + '.' : ''}${symbol.name}`
