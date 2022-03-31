import React from 'react'

import { RouteComponentProps } from 'react-router'

import { RepositoryFields } from '../../graphql-operations'
import { RepositoryReleasesTagsPage } from '../releases/RepositoryReleasesTagsPage'
import { RepoContainerContext } from '../RepoContainer'

interface Props
    extends RouteComponentProps<{}>,
        Pick<RepoContainerContext, 'repo' | 'routePrefix' | 'repoHeaderContributionsLifecycleProps'> { // TODO: do we need these props?
    repo: RepositoryFields
}

/**
 * Renders repository's tags.
 */
export const RepositoryTagTab: React.FunctionComponent<Props> = ({ repo, history, location }) => (
    <div className="repository-graph-area">
        <div className="container">
            <div className="container-inner">
                <RepositoryReleasesTagsPage repo={repo} history={history} location={location} />
            </div>
        </div>
    </div>
)
