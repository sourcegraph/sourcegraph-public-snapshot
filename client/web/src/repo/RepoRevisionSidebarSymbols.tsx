import React, { useState, useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { entries, escapeRegExp, flatMap, flow, groupBy, isEqual } from 'lodash/fp'
import { NavLink, useLocation } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useDebounce } from '@sourcegraph/wildcard'

import { useConnection } from '../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionForm,
    ConnectionContainer,
    ConnectionLoading,
    ConnectionSummary,
    ConnectionError,
    SummaryContainer,
    ShowMoreButton,
} from '../components/FilteredConnection/ui'
import { Scalars, SymbolNodeFields, SymbolsResult, SymbolsVariables } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

import styles from './RepoRevisionSidebarSymbols.module.scss'

interface SymbolNodeProps {
    node: Sym
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
    return (
        <li className={styles.repoRevisionSidebarSymbolsNode}>
            <NavLink
                to={node.url}
                isActive={isActiveFunc}
                className={classNames('test-symbol-link', styles.link)}
                activeClassName={styles.linkActive}
                onClick={onHandleClick}
            >
                <SymbolIcon kind={node.kind} className="mr-1" />
                <span className={styles.name} data-testid="symbol-name">
                    {node.name}
                </span>
            </NavLink>
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

    const { connection, error, loading, hasNextPage, fetchMore } = useConnection<
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
            {error && <ConnectionError errors={[error.message]} compact={true} />}
            {connection && (
                <HierarchicalSymbols
                    symbols={heirarchicalSymbols}
                    render={args => (
                        <SymbolNode
                            node={args.symbol}
                            onHandleClick={onHandleSymbolClick}
                            isActive={isSymbolActive(args.symbol.url)}
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
    symbols: Sym[]
    render: (props: { symbol: Sym; nestedRender: HierarchicalSymbolsProps['render'] }) => React.ReactElement
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

type Sym = SymbolNodeFields & { children: Sym[] }

const hierarchyOf = (symbols: SymbolNodeFields[]): Sym[] => {
    const fullNameToSymbol = new Map<string, SymbolNodeFields>(symbols.map(symbol => [fullName(symbol), symbol]))
    const fullNameToSym = new Map<string, Sym>()
    const topLevelSymbols: Sym[] = []

    const visit = (fullName: string): void => {
        if (fullName === '') {
            return
        }

        const symbol = fullNameToSymbol.get(fullName)
        if (!symbol) {
            return
        }

        // Sym might already exist if we've already visited a child of this symbol.
        const sym = fullNameToSym.get(fullName) || { ...symbol, children: [] }
        fullNameToSym.set(fullName, sym)

        const parentFullName = fullName.split('.').slice(0, -1).join('.')
        if (!parentFullName) {
            // No parent, add to top-level
            topLevelSymbols.push(sym)
            return
        }

        const parentSymbol = fullNameToSymbol.get(parentFullName)
        const parentSym = fullNameToSym.get(parentFullName)
        if (parentSym) {
            // Parent exists, add to parent
            parentSym.children.push(sym)
        } else if (parentSymbol) {
            // Create parent node and add current symbol to it
            fullNameToSym.set(parentFullName, { ...parentSymbol, children: [sym] })
        } else {
            // This should never happen!!
            logger.error(new Error(`Could not find parent symbol for ${fullName}`))
        }
    }

    for (const symbol of symbols) {
        visit(fullName(symbol))
    }

    // Sort everything
    for (const sym of fullNameToSym.values()) {
        sym.children.sort((a, b) => a.name.localeCompare(b.name))
    }

    return topLevelSymbols
}

const fullName = (symbol: SymbolNodeFields): string =>
    `${symbol.containerName ? symbol.containerName + '.' : ''}${symbol.name}`
