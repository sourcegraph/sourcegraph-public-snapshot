import React, { useEffect, useMemo } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { requestGraphQL } from '../../../backend/graphql'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { RepoHeaderContributionsLifecycleProps } from '../../../repo/RepoHeader'
import { eventLogger } from '../../../tracking/eventLogger'
import { DependenciesFields, DependenciesResult, DependenciesVariables } from '../../../graphql-operations'
import { RepoRevisionContainerContext } from '../../../repo/RepoRevisionContainer'

const DependenciesGQLFragment = gql`
    fragment DependenciesFields on GitTree {
        lsif {
            packages {
                nodes {
                    name
                    version
                    manager
                }
            }
        }
    }
`

const queryRepositoryPackages = (vars: DependenciesVariables): Observable<DependenciesFields | null> =>
    requestGraphQL<DependenciesResult, DependenciesVariables>(
        gql`
            query Dependencies($repo: ID!, $commitID: String!, $path: String!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $commitID) {
                            tree(path: $path) {
                                ...DependenciesFields
                            }
                        }
                    }
                }
            }
            ${DependenciesGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.commit?.tree || null)
    )

interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'resolvedRev'>,
        RepoHeaderContributionsLifecycleProps,
        BreadcrumbSetters {}

export const RepositoryDependenciesPage: React.FunctionComponent<Props> = ({ repo, resolvedRev, useBreadcrumb }) => {
    useEffect(() => {
        eventLogger.logViewEvent('RepositoryDependencies')
    }, [])

    useBreadcrumb(useMemo(() => ({ key: 'dependencies', element: <>Dependencies</> }), []))

    const data = useObservable(
        useMemo(() => queryRepositoryPackages({ repo: repo.id, commitID: resolvedRev.commitID, path: '' }), [
            repo.id,
            resolvedRev.commitID,
        ])
    )

    return (
        <div>
            <h2>Dependencies</h2>
            {data ? (
                <ul>
                    {data.lsif?.packages.nodes.map((package_, index) => (
                        <li key={index}>{JSON.stringify(package_)}</li>
                    ))}
                </ul>
            ) : (
                'Loading...'
            )}
        </div>
    )
}
