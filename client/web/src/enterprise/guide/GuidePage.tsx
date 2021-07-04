import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React, { useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { FileLocations } from '@sourcegraph/branded/src/components/panel/views/FileLocations'
import { Location } from '@sourcegraph/extension-api-types'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { memoizeObservable } from '@sourcegraph/shared/src/util/memoizeObservable'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { requestGraphQL } from '../../backend/graphql'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { GuideInfoVariables, GuidePageFields, GuideInfoResult } from '../../graphql-operations'
import { fetchHighlightedFileLineRanges } from '../../repo/backend'
import { GitCommitNode } from '../../repo/commits/GitCommitNode'
import { gitCommitFragment } from '../../repo/commits/RepositoryCommitsPage'
import { RepoHeaderContributionsLifecycleProps } from '../../repo/RepoHeader'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { eventLogger } from '../../tracking/eventLogger'

import { GuideHoverFieldsGQLFragment, GuideHover } from './GuideHover'
import { GuideViewOptionsProps, useGuideViewOptions } from './useGuideViewOptions'

const GuidePageFieldsGQLFragment = gql`
    fragment GuidePageFields on GuideInfo {
        hello
        url
        ...GuideHoverFields

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
    }
    ${GuideHoverFieldsGQLFragment}
    ${gitCommitFragment}
`
const queryGuideInfoUncached = (vars: GuideInfoVariables): Observable<GuidePageFields> =>
    requestGraphQL<GuideInfoResult, GuideInfoVariables>(
        gql`
            query GuideInfo($repository: GuideRepositoryInput!, $selections: [GuideSelectionInput!]!) {
                guideInfo(repository: $repository, selections: $selections) {
                    ...GuidePageFields
                }
            }
            ${GuidePageFieldsGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.guideInfo)
    )

const queryGuideInfo = memoizeObservable(queryGuideInfoUncached, parameters => JSON.stringify(parameters))

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
        GuideViewOptionsProps {}

export const GuidePage: React.FunctionComponent<Props> = ({
    repo,
    revision,
    resolvedRev,
    match: {
        params: { scheme, identifier },
    },
    useBreadcrumb,
    history,
    location,
    settingsCascade,
    ...props
}) => {
    useEffect(() => {
        document.body.classList.add('guide-page')
        return () => document.body.classList.remove('guide-page')
    })

    useEffect(() => {
        eventLogger.logViewEvent('Guide')
    }, [])

    const guideInfo = useObservable(
        useMemo(
            () =>
                queryGuideInfo({
                    repository: {
                        id: repo.id,
                        revision,
                        commitID: resolvedRev.commitID,
                    },
                    selections: [
                        {
                            symbolMonikers: [{ scheme, identifier }],
                        },
                    ],
                }),
            [identifier, repo.id, resolvedRev.commitID, revision, scheme]
        )
    )

    useBreadcrumb(
        useMemo(
            () =>
                guideInfo === null
                    ? null
                    : {
                          key: 'symbol/current',
                          element: guideInfo ? (
                              <Link to={guideInfo.url}>{guideInfo.hello}</Link>
                          ) : (
                              <LoadingSpinner className="icon-inline" />
                          ),
                      },
            [guideInfo]
        )
    )

    return guideInfo === null ? (
        <p className="p-3 text-muted h3">Not found</p>
    ) : guideInfo === undefined ? (
        <LoadingSpinner className="m-3" />
    ) : (
        <>
            <GuideHover
                {...props}
                guideInfo={guideInfo}
                afterSignature={
                    <div className="d-flex align-items-center mx-3">
                        {/* <SymbolActions guideInfo={guideInfo} /> */}
                    </div>
                }
                className="mx-3 mt-3"
            />
            {guideInfo.references.nodes.length > 1 && (
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
                            guideInfo.references.nodes
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
                        isLightTheme={false /* TODO(sqs) */}
                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                        settingsCascade={settingsCascade}
                        versionContext={undefined /* TODO(sqs) */}
                    />
                </section>
            )}
            {guideInfo.editCommits && guideInfo.editCommits.nodes.length > 0 && (
                <section id="refs" className="my-4">
                    <h2 className="mt-0 mx-3 mb-0 h4">Changes</h2>
                    {guideInfo.editCommits.nodes.map(commit => (
                        <GitCommitNode key={commit.oid} node={commit} className="px-3" compact={true} />
                    ))}
                </section>
            )}
            {/* {symbol.children.nodes.length > 0 && (
                <ContainerSymbolsList
                    symbols={symbol.children.nodes.sort((a, b) => (a.kind < b.kind ? -1 : 1))}
                    history={history}
                />
            )} */}
        </>
    )
}
