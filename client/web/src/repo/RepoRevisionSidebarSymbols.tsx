import React, { useState, useMemo, Suspense } from 'react'

import classNames from 'classnames'
import { escapeRegExp, groupBy } from 'lodash'

import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import type { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Alert, useDebounce, ErrorMessage } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionForm,
    ConnectionContainer,
    ConnectionLoading,
    ConnectionSummary,
    SummaryContainer,
    ShowMoreButton,
} from '../components/FilteredConnection/ui'
import type { Scalars, SymbolNodeFields, SymbolsResult, SymbolsVariables } from '../graphql-operations'

import { RepoRevisionSidebarSymbolTree } from './RepoRevisionSidebarSymbolTree'

import styles from './RepoRevisionSidebarSymbols.module.scss'
import { SymbolWithChildren, hierarchyOf } from './utils'

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
    focusKey?: string
}

export const RepoRevisionSidebarSymbols: React.FunctionComponent<
    React.PropsWithChildren<RepoRevisionSidebarSymbolsProps>
> = ({ repoID, revision = '', activePath, focusKey, onHandleSymbolClick }) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)

    // URL is the most unique part we have about a symbol node. We use it
    // instead of the index to avoid pointing to the wrong symbol when the data
    // changes.
    const [selectedSymbolUrl, setSelectedSymbolUrl] = useState<null | string>(null)

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

    const hierarchicalSymbols = useMemo<SymbolWithChildren[]>(
        () =>
            Object.values(groupBy(connection?.nodes ?? [], symbol => symbol.location.resource.path)).flatMap(symbols =>
                hierarchyOf(symbols)
            ),
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
            {connection && !loading ? (
                <Suspense fallback={<ConnectionLoading compact={true} />}>
                    <RepoRevisionSidebarSymbolTree
                        // We throw away the component state whenever the underlying query
                        // data changes to avoid complicated bookkeeping in the tree
                        // component.
                        key={activePath + ':' + query}
                        focusKey={focusKey}
                        selectedSymbolUrl={selectedSymbolUrl}
                        setSelectedSymbolUrl={setSelectedSymbolUrl}
                        symbols={hierarchicalSymbols}
                        onClick={onHandleSymbolClick}
                    />
                </Suspense>
            ) : null}
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
