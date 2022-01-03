import classNames from 'classnames'
import * as H from 'history'
import { escapeRegExp, isEqual } from 'lodash'
import * as React from 'react'
import { useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'

import { gql, dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionForm,
    ConnectionList,
    ConnectionContainer,
    ConnectionLoading,
    ConnectionSummary,
    ConnectionError,
    SummaryContainer,
    ShowMoreButton,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { useDebounce, Tooltip } from '@sourcegraph/wildcard'

import { Scalars, SymbolNodeFields, SymbolsResult, SymbolsVariables } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

import styles from './RepoRevisionSidebarSymbols.module.scss'

function symbolIsActive(symbolLocation: string, currentLocation: H.Location): boolean {
    const current = parseBrowserRepoURL(H.createPath(currentLocation))
    const symbol = parseBrowserRepoURL(symbolLocation)
    return (
        current.repoName === symbol.repoName &&
        current.revision === symbol.revision &&
        current.filePath === symbol.filePath &&
        isEqual(current.position, symbol.position)
    )
}

const symbolIsActiveTrue = (): boolean => true
const symbolIsActiveFalse = (): boolean => false

interface SymbolNodeProps {
    node: SymbolNodeFields
    location: H.Location
    onHandleClick: () => void
}

const SymbolNode: React.FunctionComponent<SymbolNodeProps> = ({ node, location, onHandleClick }) => {
    const isActiveFunc = symbolIsActive(node.url, location) ? symbolIsActiveTrue : symbolIsActiveFalse
    return (
        <li className={styles.repoRevisionSidebarSymbolsNode} data-tooltip={node.location.resource.path}>
            <NavLink
                to={node.url}
                isActive={isActiveFunc}
                className={classNames('test-symbol-link', styles.link)}
                activeClassName={styles.linkActive}
                onClick={onHandleClick}
            >
                <SymbolIcon kind={node.kind} className="icon-inline mr-1 test-symbol-icon" />
                <span className={classNames('test-symbol-name', styles.name)}>{node.name}</span>
                {node.containerName && (
                    <span className={styles.containerName}>
                        <small>{node.containerName}</small>
                    </span>
                )}
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

export const RepoRevisionSidebarSymbols: React.FunctionComponent<RepoRevisionSidebarSymbolsProps> = ({
    repoID,
    revision = '',
    activePath,
    onHandleSymbolClick,
}) => {
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
            includePatterns: [escapeRegExp(activePath)],
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                throw new Error(`Node ${repoID} not found`)
            }
            if (node.__typename !== 'Repository') {
                throw new Error(`Node is a ${node.__typename}, not a Repository`)
            }
            if (!node.commit?.symbols?.nodes) {
                throw new Error('Could not resolve commit symbols for repository')
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

    return (
        <ConnectionContainer className={classNames('h-100', styles.repoRevisionSidebarSymbols)} compact={true}>
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
            {error && <ConnectionError errors={[error.message]} compact={true} />}
            {connection && (
                <ConnectionList compact={true}>
                    <Tooltip />
                    {connection.nodes.map((node, index) => (
                        <SymbolNode key={index} node={node} location={location} onHandleClick={onHandleSymbolClick} />
                    ))}
                </ConnectionList>
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
