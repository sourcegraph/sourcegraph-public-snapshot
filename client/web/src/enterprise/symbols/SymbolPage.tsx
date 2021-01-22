import React, { useEffect, useMemo, useState } from 'react'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { requestGraphQL } from '../../backend/graphql'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepoHeaderContributionsLifecycleProps } from '../../repo/RepoHeader'
import { eventLogger } from '../../tracking/eventLogger'
import {
    RepositoryExpSymbolResult,
    RepositoryExpSymbolVariables,
    SymbolPageSymbolFields,
} from '../../graphql-operations'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { RouteComponentProps } from 'react-router'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { SymbolHoverGQLFragment, SymbolHover } from './SymbolHover'
import { SymbolsSidebarOptionsSetterProps } from './SymbolsArea'
import { SymbolsViewOptionsProps } from './useSymbolsViewOptions'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ContainerSymbolsList, ContainerSymbolsListSymbolGQLFragment } from './ContainerSymbolsList'
import { Link } from 'react-router-dom'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { makeRepoURI } from '../../../../shared/src/util/url'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import { fetchHighlightedFileLineRanges } from '../../repo/backend'
import { Location } from '@sourcegraph/extension-api-types'
import { gitCommitFragment } from '../../repo/commits/RepositoryCommitsPage'
import { GitCommitNode } from '../../repo/commits/GitCommitNode'
import { FileLocations } from '../../../../branded/src/components/panel/views/FileLocations'
import { SymbolsSidebarContainerSymbolGQLFragment } from './SymbolsSidebar'
import { SymbolActions, SymbolActionsGQLFragment } from './SymbolActions'
import { SymbolStatsSummary } from './SymbolStatsSummary'
import { VersionContextProps } from '../../../../shared/src/search/util'

const SymbolPageSymbolGQLFragment = gql`
    fragment SymbolPageSymbolFields on ExpSymbol {
        text
        detail
        kind
        url
        ...SymbolHoverFields
        ...SymbolActionsFields
        references {
            nodes {
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
                resource {
                    path
                    commit {
                        oid
                    }
                    repository {
                        name
                    }
                }
            }
        }
        editCommits {
            nodes {
                ...GitCommitFields
            }
        }

        rootAncestor {
            ...SymbolsSidebarContainerSymbolFields
        }
        children(filters: $filters) {
            nodes {
                ...ContainerSymbolsListSymbolFields
            }
        }
    }
    ${SymbolHoverGQLFragment}
    ${SymbolActionsGQLFragment}
    ${gitCommitFragment}
    ${SymbolsSidebarContainerSymbolGQLFragment}
    ${ContainerSymbolsListSymbolGQLFragment}
`

const queryRepositorySymbolUncached = (vars: RepositoryExpSymbolVariables): Observable<SymbolPageSymbolFields | null> =>
    requestGraphQL<RepositoryExpSymbolResult, RepositoryExpSymbolVariables>(
        gql`
            query RepositoryExpSymbol(
                $repo: ID!
                $revision: String!
                $moniker: MonikerInput!
                $filters: SymbolFilters!
            ) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $revision) {
                            tree(path: "") {
                                expSymbol(moniker: $moniker) {
                                    ...SymbolPageSymbolFields
                                }
                            }
                        }
                    }
                }
            }
            ${SymbolPageSymbolGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.commit?.tree?.expSymbol || null)
    )

const queryRepositorySymbol = memoizeObservable(queryRepositorySymbolUncached, parameters => JSON.stringify(parameters))

export interface SymbolRouteProps {
    scheme: string
    identifier: string
}

interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'resolvedRev' | 'revision'>,
    RouteComponentProps<SymbolRouteProps>,
    RepoHeaderContributionsLifecycleProps,
    BreadcrumbSetters,
    SettingsCascadeProps,
    SymbolsSidebarOptionsSetterProps,
    SymbolsViewOptionsProps,
    VersionContextProps {
    isLightTheme: boolean
}

export const SymbolPage: React.FunctionComponent<Props> = ({
    repo,
    revision,
    resolvedRev,
    viewOptions,
    setSidebarOptions,
    match: {
        params: { scheme, identifier },
    },
    useBreadcrumb,
    history,
    location,
    settingsCascade,
    isLightTheme,
    ...props
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('Symbol')
    }, [])

    const symbol = useObservable(
        useMemo(
            () =>
                queryRepositorySymbol({
                    repo: repo.id,
                    revision,
                    moniker: { scheme, identifier },
                    filters: { internals: viewOptions.internals },
                }),
            [identifier, repo.id, revision, scheme, viewOptions]
        )
    )

    // Cache the rootAncestor since it doesn't change across navigations. TODO(sqs): it sometimes
    // can, eg if there is a direct link, so figure out a way to make this correct.
    const [cachedRootAncestor, setCachedRootAncestor] = useState<SymbolPageSymbolFields['rootAncestor']>()
    useEffect(() => {
        if (symbol) {
            setCachedRootAncestor(symbol.rootAncestor)
        }
    }, [setCachedRootAncestor, symbol])

    useBreadcrumb = useBreadcrumb(
        useMemo(
            () =>
                cachedRootAncestor
                    ? {
                        key: 'symbol/container',
                        element: <Link to={cachedRootAncestor.url}>{cachedRootAncestor.text}</Link>,
                    }
                    : null,
            [cachedRootAncestor]
        )
    ).useBreadcrumb
    useBreadcrumb(
        useMemo(
            () =>
                symbol === null
                    ? null
                    : {
                        key: 'symbol/current',
                        element: symbol ? (
                            <Link to={symbol.url}>{symbol.text}</Link>
                        ) : (
                                <LoadingSpinner className="icon-inline" />
                            ),
                    },
            [symbol]
        )
    )

    useEffect(() => setSidebarOptions(cachedRootAncestor ? { containerSymbol: cachedRootAncestor } : null), [
        symbol,
        setSidebarOptions,
        cachedRootAncestor,
    ])

    return symbol === null ? (
        <p className="p-3 text-muted h3">Not found</p>
    ) : symbol === undefined ? (
        <LoadingSpinner className="m-3" />
    ) : (
                <>
                    <SymbolHover
                        {...props}
                        symbol={symbol}
                        afterSignature={
                            <div className="d-flex align-items-center mx-3">
                                <SymbolActions symbol={symbol} />
                                <SymbolStatsSummary symbol={symbol} className="text-muted ml-3 small" />
                            </div>
                        }
                        className="mx-3 mt-3"
                        history={history}
                        location={location}
                    />
                    {symbol.references.nodes.length > 1 && (
                        <section id="refs" className="mt-2">
                            <h2 className="mt-0 mx-3 mb-0 h4">Examples</h2>
                            <style>
                                {
                                    'td.line { display: none; } .code-excerpt .code { padding-left: 0.25rem !important; } .result-container__header { display: none; } .result-container { border: solid 1px var(--border-color) !important; border-width: 1px !important; margin: 1rem; }'
                                }
                            </style>
                            <FileLocations
                                location={location}
                                locations={of(
                                    symbol.references.nodes
                                        .slice(0, -1)
                                        .slice(0, 3)
                                        .map<Location>(reference => ({
                                            uri: makeRepoURI({
                                                repoName: reference.resource.repository.name,
                                                commitID: reference.resource.commit.oid,
                                                filePath: reference.resource.path,
                                            }),
                                            range: reference.range!,
                                        }))
                                )}
                                icon={SourceRepositoryIcon}
                                isLightTheme={isLightTheme}
                                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                settingsCascade={settingsCascade}
                                versionContext={props.versionContext}
                            />
                        </section>
                    )}
                    {symbol.editCommits && symbol.editCommits.nodes.length > 0 && (
                        <section id="refs" className="my-4">
                            <h2 className="mt-0 mx-3 mb-0 h4">Changes</h2>
                            {symbol.editCommits.nodes.map(commit => (
                                <GitCommitNode key={commit.oid} node={commit} className="px-3" compact={true} />
                            ))}
                        </section>
                    )}
                    {symbol.children.nodes.length > 0 && (
                        <ContainerSymbolsList
                            symbols={symbol.children.nodes.sort((a, b) => (a.kind < b.kind ? -1 : 1))}
                            history={history}
                        />
                    )}
                </>
            )
}
