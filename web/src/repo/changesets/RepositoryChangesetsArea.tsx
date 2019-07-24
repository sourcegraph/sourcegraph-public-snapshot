import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { pluralize } from '../../../../shared/src/util/strings'
import { RepoContainerContext } from '../RepoContainer'
import { RepoHeaderBreadcrumbNavItem } from '../RepoHeaderBreadcrumbNavItem'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'
import { useRepositoryChangesets } from './useRepositoryChangesets'

interface Props
    extends RouteComponentProps<{}>,
        Pick<RepoContainerContext, 'repo' | 'routePrefix' | 'repoHeaderContributionsLifecycleProps'> {
    repo: GQL.IRepository
}

const LOADING: 'loading' = 'loading'

/**
 * A list of the repository's changesets.
 */
export const RepositoryChangesetsArea: React.FunctionComponent<Props> = ({
    repo,
    repoHeaderContributionsLifecycleProps,
}) => {
    const changesetsOrError = useRepositoryChangesets(repo)
    return (
        <div className="repository-changesets-area">
            <RepoHeaderContributionPortal
                position="nav"
                element={<RepoHeaderBreadcrumbNavItem key="changesets">Changesets</RepoHeaderBreadcrumbNavItem>}
                repoHeaderContributionsLifecycleProps={repoHeaderContributionsLifecycleProps}
            />
            <div className="container mt-4">
                {changesetsOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline mt-3" />
                ) : isErrorLike(changesetsOrError) ? (
                    <div className="alert alert-danger mt-3">{changesetsOrError.message}</div>
                ) : (
                    <div className="card">
                        <div className="card-header">
                            <span className="text-muted">
                                {changesetsOrError.totalCount} {pluralize('changeset', changesetsOrError.totalCount)}
                            </span>
                        </div>
                        {changesetsOrError.nodes.length > 0 ? (
                            <ul className="list-group list-group-flush">
                                {changesetsOrError.nodes.map((changeset, i) => (
                                    <li key={i} className="list-group-item position-relative">
                                        <LinkOrSpan to={changeset.externalURL} className="stretched-link">
                                            {changeset.title}
                                        </LinkOrSpan>
                                    </li>
                                ))}
                            </ul>
                        ) : (
                            <div className="p-2 text-muted">No changesets.</div>
                        )}
                    </div>
                )}
            </div>
        </div>
    )
}
