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
import { useRepositoryThreads } from './useRepositoryThreads'

interface Props
    extends RouteComponentProps<{}>,
        Pick<RepoContainerContext, 'repo' | 'routePrefix' | 'repoHeaderContributionsLifecycleProps'> {
    repo: GQL.IRepository
}

const LOADING: 'loading' = 'loading'

/**
 * A list of the repository's threads.
 */
export const RepositoryThreadsArea: React.FunctionComponent<Props> = ({
    repo,
    repoHeaderContributionsLifecycleProps,
}) => {
    const threadsOrError = useRepositoryThreads(repo)
    return (
        <div className="repository-threads-area">
            <RepoHeaderContributionPortal
                position="nav"
                element={<RepoHeaderBreadcrumbNavItem key="threads">Pull requests</RepoHeaderBreadcrumbNavItem>}
                repoHeaderContributionsLifecycleProps={repoHeaderContributionsLifecycleProps}
            />
            <div className="container mt-4">
                {threadsOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline mt-3" />
                ) : isErrorLike(threadsOrError) ? (
                    <div className="alert alert-danger mt-3">{threadsOrError.message}</div>
                ) : (
                    <div className="card">
                        <div className="card-header">
                            <span className="text-muted">
                                {threadsOrError.totalCount} {pluralize('pull request', threadsOrError.totalCount)}
                            </span>
                        </div>
                        {threadsOrError.nodes.length > 0 ? (
                            <ul className="list-group list-group-flush">
                                {threadsOrError.nodes.map((thread, i) => (
                                    <li key={i} className="list-group-item position-relative">
                                        <LinkOrSpan to={thread.externalURL} className="stretched-link">
                                            {thread.title}
                                        </LinkOrSpan>
                                    </li>
                                ))}
                            </ul>
                        ) : (
                            <div className="p-2 text-muted">No threads.</div>
                        )}
                    </div>
                )}
            </div>
        </div>
    )
}
