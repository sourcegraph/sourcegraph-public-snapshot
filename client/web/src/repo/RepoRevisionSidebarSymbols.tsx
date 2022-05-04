import * as React from 'react'
import { useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { escapeRegExp, isEqual } from 'lodash'
import { NavLink, useLocation } from 'react-router-dom'

import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useDebounce } from '@sourcegraph/wildcard'

import { useConnection } from '../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionForm,
    ConnectionList,
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
    node: SymbolNodeFields
    onHandleClick: () => void
    isActive: boolean
}

const SymbolNode: React.FunctionComponent<React.PropsWithChildren<SymbolNodeProps>> = ({
    node,
    onHandleClick,
    isActive,
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
                <SymbolIcon kind={node.kind} className="mr-1 test-symbol-icon" />
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
                    {connection.nodes.map((node, index) => (
                        <SymbolNode
                            key={index}
                            node={node}
                            onHandleClick={onHandleSymbolClick}
                            isActive={isSymbolActive(node.url)}
                        />
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
