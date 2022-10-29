import * as React from 'react'
import { useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { sortBy } from 'lodash'
import { entries, escapeRegExp, flatMap, flow, groupBy, isEqual, map } from 'lodash/fp'
import { NavLink, useLocation } from 'react-router-dom'

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
import { Scalars, SymbolKind, SymbolNodeFields, SymbolsResult, SymbolsVariables } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

import styles from './RepoRevisionSidebarSymbols.module.scss'

interface SymbolNodeProps {
    node: SymbolNodeFields
    onHandleClick: () => void
    isActive: boolean
    style: React.CSSProperties
}

const SymbolNode: React.FunctionComponent<React.PropsWithChildren<SymbolNodeProps>> = ({
    node,
    onHandleClick,
    isActive,
    style,
}) => {
    const isActiveFunc = (): boolean => isActive
    return (
        // eslint-disable-next-line react/forbid-dom-props
        <li className={styles.repoRevisionSidebarSymbolsNode} style={style}>
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
                    symbols={connection.nodes}
                    render={args =>
                        typeof args.symbol === 'string' ? (
                            // eslint-disable-next-line react/forbid-dom-props
                            <li className={styles.repoRevisionSidebarSymbolsNode} style={padding(args.symbol)}>
                                <span className={styles.link}>
                                    <SymbolIcon kind={SymbolKind.UNKNOWN} className="mr-1" />
                                    {args.symbol}
                                </span>
                            </li>
                        ) : (
                            <SymbolNode
                                node={args.symbol}
                                onHandleClick={onHandleSymbolClick}
                                isActive={isSymbolActive(args.symbol.url)}
                                style={padding(args.symbol)}
                            />
                        )
                    }
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
    symbols: SymbolNodeFields[]
    render: (props: { symbol: SymbolNodeFields | string }) => React.ReactElement
}

const HierarchicalSymbols: React.FunctionComponent<HierarchicalSymbolsProps> = props => (
    <ul className={styles.hierarchicalSymbolsContainer}>
        {flow(
            groupBy<SymbolNodeFields>(symbol => symbol.location.resource.path),
            entries,
            flatMap(([, symbols]) => hierarchyOf(symbols)),
            map(symbol => <props.render key={typeof symbol === 'string' ? symbol : symbol.url} symbol={symbol} />)
        )(props.symbols)}
    </ul>
)

const hierarchyOf = (symbols: SymbolNodeFields[]): (SymbolNodeFields | string)[] => {
    const map1 = new Map<string, SymbolNodeFields | string>(
        symbols.map(symbol => [`${symbol.containerName ? symbol.containerName + '.' : ''}${symbol.name}`, symbol])
    )

    for (const missing of symbols
        .filter(symbol => symbol.containerName)
        .map(symbol => symbol.containerName ?? '')
        .filter(containerName => !map1.has(containerName))) {
        map1.set(missing, missing)
    }

    return sortBy([...map1.entries()], ([fullName]) => fullName).map(([, symbol]) => symbol)
}

const depthOf = (symbol: SymbolNodeFields | string): number =>
    (typeof symbol === 'string' ? symbol : fullName(symbol)).split('.').length - 1

const fullName = (symbol: SymbolNodeFields): string =>
    `${symbol.containerName ? symbol.containerName + '.' : ''}${symbol.name}`

const padding = (symbol: SymbolNodeFields | string): React.CSSProperties => ({ paddingLeft: `${depthOf(symbol)}rem` })
