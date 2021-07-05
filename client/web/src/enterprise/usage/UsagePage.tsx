import React, { useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { memoizeObservable } from '@sourcegraph/shared/src/util/memoizeObservable'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { requestGraphQL } from '../../backend/graphql'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { UsagePageVariables, UsagePageFields, UsagePageResult } from '../../graphql-operations'
import { RepoHeaderContributionsLifecycleProps } from '../../repo/RepoHeader'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { eventLogger } from '../../tracking/eventLogger'

import { SymbolReferenceGroupsSection, SymbolReferenceGroupGQLFragment } from './symbol/SymbolReferenceGroupsSection'

const UsagePageFieldsGQLFragment = gql`
    fragment UsagePageFields on ExpSymbol {
        text
        url

        usage {
            referenceGroups {
                ...SymbolReferenceGroup
            }
        }
    }
    ${SymbolReferenceGroupGQLFragment}
`
const queryUsagePageUncached = (vars: UsagePageVariables): Observable<UsagePageFields | null> =>
    requestGraphQL<UsagePageResult, UsagePageVariables>(
        gql`
            query UsagePage($repo: ID!, $commitID: String!, $inputRevspec: String!, $moniker: MonikerInput!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $commitID, inputRevspec: $inputRevspec) {
                            tree(path: "/") {
                                expSymbol(moniker: $moniker) {
                                    ...UsagePageFields
                                }
                            }
                        }
                    }
                }
            }
            ${UsagePageFieldsGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.commit?.tree?.expSymbol || null)
    )

const queryUsagePage = memoizeObservable(queryUsagePageUncached, parameters => JSON.stringify(parameters))

export interface UsageRouteProps {
    scheme: string
    identifier: string
}

interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'resolvedRev' | 'revision'>,
        RouteComponentProps<UsageRouteProps>,
        RepoHeaderContributionsLifecycleProps,
        BreadcrumbSetters,
        SettingsCascadeProps,
        ThemeProps,
        VersionContextProps {}

export const UsagePage: React.FunctionComponent<Props> = ({
    repo,
    revision,
    resolvedRev,
    match: {
        params: { scheme, identifier },
    },
    useBreadcrumb,
    history,
    ...props
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('Usage')
    }, [])

    const symbol = useObservable(
        useMemo(
            () =>
                queryUsagePage({
                    repo: repo.id,
                    commitID: resolvedRev.commitID,
                    inputRevspec: revision,
                    moniker: { scheme, identifier },
                }),
            [identifier, repo.id, resolvedRev.commitID, revision, scheme]
        )
    )

    useBreadcrumb(
        useMemo(
            () =>
                symbol === null
                    ? null
                    : {
                          key: 'usage',
                          element: symbol ? (
                              <>
                                  Usage: <Link to={symbol.url}>{symbol.text}</Link>
                              </>
                          ) : (
                              <LoadingSpinner className="icon-inline" />
                          ),
                      },
            [symbol]
        )
    )

    /*     const onClick = useCallback<React.MouseEventHandler>(e => {
        window.parent.postMessage({ type: 'usageClick' }, '*')
    }, [])
 */
    return symbol === null ? (
        <p className="p-3 text-muted h3">Not found</p>
    ) : symbol === undefined ? (
        <LoadingSpinner className="m-3" />
    ) : (
        <>
            {symbol.usage.referenceGroups.length > 0 && (
                <section>
                    <SymbolReferenceGroupsSection referenceGroups={symbol.usage.referenceGroups} {...props} />
                </section>
            )}
        </>
    )
}
