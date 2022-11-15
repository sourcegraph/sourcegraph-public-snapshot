import * as React from 'react'
import { useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { sortBy } from 'lodash'
import { entries, escapeRegExp, flatMap, flow, groupBy, isEqual, map } from 'lodash/fp'
import { NavLink, useLocation } from 'react-router-dom'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Alert, useDebounce } from '@sourcegraph/wildcard'

import { useConnection } from '../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionForm,
    ConnectionContainer,
    ConnectionLoading,
    ConnectionSummary,
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
            {error && (
                <Alert variant={error.message.includes('Estimated completion') ? 'info' : 'danger'}>
                    <ErrorMessage error={error.message} />
                </Alert>
            )}
            {connection && (
                <HierarchicalSymbols
                    symbols={connection.nodes}
                    render={args =>
                        args.symbol.__typename === 'IntermediateSymbol' ? (
                            // eslint-disable-next-line react/forbid-dom-props
                            <li className={styles.repoRevisionSidebarSymbolsNode} style={padding(args.symbol)}>
                                <span className={styles.link}>
                                    <SymbolIcon kind={SymbolKind.UNKNOWN} className="mr-1" />
                                    {args.symbol.name}
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
    render: (props: { symbol: Sym }) => React.ReactElement
}

const HierarchicalSymbols: React.FunctionComponent<HierarchicalSymbolsProps> = props => (
    <ul className={styles.hierarchicalSymbolsContainer}>
        {flow(
            groupBy<SymbolNodeFields>(symbol => symbol.location.resource.path),
            entries,
            flatMap(([, symbols]) => hierarchyOf(symbols)),
            map(symbol => <props.render key={'url' in symbol ? symbol.url : fullName(symbol)} symbol={symbol} />)
        )(props.symbols)}
    </ul>
)

interface IntermediateSymbol {
    __typename: 'IntermediateSymbol'
    name: string
    language: string
}

type Sym = (SymbolNodeFields | IntermediateSymbol) & { containers: string[] }

const hierarchyOf = (symbols: SymbolNodeFields[]): Sym[] => {
    const fullNameToSymbol = new Map<string, SymbolNodeFields>(symbols.map(symbol => [fullName(symbol), symbol]))
    const fullNameToSym = new Map<string, Sym>()

    const visit = (fullName: string, language: string): string[] => {
        if (fullName === '') {
            return []
        }

        const sym = fullNameToSym.get(fullName)
        if (sym) {
            return [...sym.containers, sym.name]
        }

        const symbol = fullNameToSymbol.get(fullName)
        if (symbol) {
            const containers = symbol.containerName ? visit(fullName.split('.').slice(0, -1).join('.'), language) : []
            fullNameToSym.set(fullName, { ...symbol, containers })
            return [...containers, symbol.name]
        }

        const containers = visit(fullName.split('.').slice(0, -1).join('.'), language)
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        const name = fullName.split('.').pop()!
        fullNameToSym.set(fullName, {
            __typename: 'IntermediateSymbol',
            containers,
            name,
            language,
        })
        return [...containers, name]
    }

    for (const symbol of symbols) {
        visit(fullName(symbol), symbol.language)
    }

    return sortBy([...fullNameToSym.entries()], ([fullName]) => fullName).map(([, symbol]) => symbol)
}

const fullName = (symbol: SymbolNodeFields | Sym): string => {
    if ('containers' in symbol) {
        return [...symbol.containers, symbol.name].join('.')
    }
    return `${symbol.containerName ? symbol.containerName + '.' : ''}${symbol.name}`
}

const padding = (symbol: Sym): React.CSSProperties => ({ paddingLeft: `${symbol.containers.length}rem` })
