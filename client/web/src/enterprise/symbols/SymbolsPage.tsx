import React, { useEffect, useMemo } from 'react'
import H from 'history'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { requestGraphQL } from '../../backend/graphql'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepoHeaderContributionsLifecycleProps } from '../../repo/RepoHeader'
import { eventLogger } from '../../tracking/eventLogger'
import {
    RepositoryExpSymbolsFields,
    RepositoryExpSymbolsVariables,
    RepositoryExpSymbolsResult,
} from '../../graphql-operations'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { SymbolHoverGQLFragment } from './SymbolHover'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { ContainerSymbolsList } from './ContainerSymbolsList'
import { SymbolsSidebarOptionsSetterProps } from './SymbolsArea'
import { SymbolsViewOptionsProps } from './useSymbolsViewOptions'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

const RepositoryExpSymbolsGQLFragment = gql`
    fragment RepositoryExpSymbolsFields on ExpSymbol {
        text
        detail
        kind
        monikers {
            identifier
        }
        url
        children {
            nodes {
                ...SymbolHoverFields
                children {
                    nodes {
                        ...SymbolHoverFields
                    }
                }
            }
        }
        ...SymbolHoverFields
    }
    ${SymbolHoverGQLFragment}
`

const queryRepositorySymbols = (vars: RepositoryExpSymbolsVariables): Observable<RepositoryExpSymbolsFields[] | null> =>
    requestGraphQL<RepositoryExpSymbolsResult, RepositoryExpSymbolsVariables>(
        gql`
            query RepositoryExpSymbols($repo: ID!, $commitID: String!, $path: String!, $filters: SymbolFilters!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $commitID) {
                            tree(path: $path) {
                                expSymbols(filters: $filters) {
                                    nodes {
                                        ...RepositoryExpSymbolsFields
                                    }
                                }
                            }
                        }
                    }
                }
            }
            ${RepositoryExpSymbolsGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.commit?.tree?.expSymbols?.nodes || null)
    )

interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'resolvedRev'>,
        SettingsCascadeProps,
        SymbolsSidebarOptionsSetterProps,
        SymbolsViewOptionsProps {
    history: H.History
    location: H.Location
}

export const SymbolsPage: React.FunctionComponent<Props> = ({ repo, resolvedRev, viewOptions, history, ...props }) => {
    useEffect(() => {
        eventLogger.logViewEvent('Symbols')
    }, [])

    const data = useObservable(
        useMemo(
            () =>
                queryRepositorySymbols({
                    repo: repo.id,
                    commitID: resolvedRev.commitID,
                    path: '.',
                    filters: viewOptions,
                }),
            [repo.id, resolvedRev.commitID, viewOptions]
        )
    )

    return data ? <ContainerSymbolsList symbols={data} history={history} /> : <LoadingSpinner className="m-3" />
}
