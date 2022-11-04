import React from 'react'

import H from 'history'

import { RepositoryFields } from '../../graphql-operations'
import { RepositoryReleasesTagsPage } from '../releases/RepositoryReleasesTagsPage'

interface Props {
    repo: RepositoryFields | undefined
    history: H.History
    location: H.Location
}

/**
 * Renders repository's tags.
 */
export const RepositoryTagTab: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repo,
    history,
    location,
}) => (
    <div className="repository-graph-area">
        <div className="container">
            <div className="container-inner">
                <RepositoryReleasesTagsPage repo={repo} history={history} location={location} />
            </div>
        </div>
    </div>
)
