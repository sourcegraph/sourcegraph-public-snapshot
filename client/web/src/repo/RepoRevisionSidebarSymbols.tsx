import React, { Suspense, useMemo, useState } from 'react'

import classNames from 'classnames'
import { escapeRegExp, groupBy } from 'lodash'

import { logger } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import type { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Alert, ErrorMessage, useDebounce } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionForm,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../components/FilteredConnection/ui'
import type { Scalars, SymbolNodeFields, SymbolsResult, SymbolsVariables } from '../graphql-operations'

import { RepoRevisionSidebarSymbolTree } from './RepoRevisionSidebarSymbolTree'

import styles from './RepoRevisionSidebarSymbols.module.scss'

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
            pageSize: BATCH_COUNT,
        },
    })

    const summary = connection && (
        <ConnectionSummary
            connection={connection}
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

// When searching symbols, results may contain only child symbols without their parents
// (e.g. when searching for "bar", a class named "Foo" with a method named "bar" will
// return "bar" as a result, and "bar" will say that "Foo" is its parent).
// The placeholder symbols exist to show the hierarchy of the results, but these placeholders
// are not interactive (cannot be clicked to navigate) and don't have any other information.
export interface SymbolPlaceholder {
    __typename: 'SymbolPlaceholder'
    name: string
}

export type SymbolWithChildren = (SymbolNodeFields | SymbolPlaceholder) & { children: SymbolWithChildren[] }

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
