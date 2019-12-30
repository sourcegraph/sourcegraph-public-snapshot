import React, { useMemo } from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { Observable } from 'rxjs'
import { queryGraphQL } from '../backend/graphql'
import { gql, dataOrThrowErrors } from '../../../shared/src/graphql/graphql'
import { map } from 'rxjs/operators'
import { useObservable } from '../util/useObservable'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { RepoLink } from '../../../shared/src/components/RepoLink'
import { Timestamp } from '../components/time/Timestamp'
import { PersonLink } from '../person/PersonLink'
import { Link } from 'react-router-dom'
import { buildSearchURLQuery } from '../../../shared/src/util/url'
import { Settings } from '../schema/settings.schema'

/**
 * Queries for a list of repositories.
 *
 * @returns Observable that emits the list of repositories
 */
const queryRepositories = (
    vars: Pick<GQL.IRepositoriesOnQueryArguments, 'names'>
): Observable<GQL.IRepositoryConnection> =>
    queryGraphQL(
        gql`
            query Repositories($first: Int, $names: [String!]) {
                repositories(first: $first, names: $names) {
                    nodes {
                        id
                        name
                        description
                        defaultBranch {
                            target {
                                commit {
                                    message
                                    author {
                                        person {
                                            displayName
                                        }
                                        date
                                    }
                                    committer {
                                        person {
                                            displayName
                                        }
                                        date
                                    }
                                }
                            }
                        }
                        createdAt
                        viewerCanAdminister
                        url
                        mirrorInfo {
                            cloned
                            cloneInProgress
                            updatedAt
                        }
                    }
                }
            }
        `,
        {
            ...vars,
            first: 5,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repositories)
    )

interface Props {
    settings: Settings
}

/**
 * A list of repositories for the homepage.
 */
export const HomeRepositories: React.FunctionComponent<Props> = ({ settings }) => {
    const repositories = useObservable(
        useMemo(
            () =>
                queryRepositories({
                    names: settings['search.repositoryGroups']?.home || null,
                }),
            [settings]
        )
    )
    return (
        <div className="card w-100">
            <h3 className="card-header d-flex align-items-center">
                Repositories {repositories === undefined && <LoadingSpinner className="icon-inline ml-2" />}
            </h3>
            {repositories !== undefined && (
                <>
                    <table className="table mb-0">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Last commit</th>
                            </tr>
                        </thead>
                        <tbody>
                            {repositories.nodes.map(repo => (
                                <tr key={repo.id}>
                                    <td className="text-nowrap">
                                        <RepoLink repoName={repo.name} to={repo.url} />
                                    </td>
                                    <td>
                                        {repo.defaultBranch?.target?.commit && (
                                            <>
                                                <Timestamp date={repo.defaultBranch?.target?.commit.author.date} /> by{' '}
                                                <PersonLink person={repo.defaultBranch?.target?.commit.author.person} />
                                            </>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                    <footer className="card-footer small py-1">
                        <Link to={`/search?${buildSearchURLQuery('repo:', GQL.SearchPatternType.literal)}`}>
                            View all repositories
                        </Link>
                    </footer>
                </>
            )}
        </div>
    )
}
