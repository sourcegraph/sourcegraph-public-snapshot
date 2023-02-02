import React from 'react'

import { RepositoryFields } from '../../graphql-operations'
import { RepositoryReleasesTagsPage } from '../releases/RepositoryReleasesTagsPage'

interface Props {
    repo: RepositoryFields | undefined
}

/**
 * Renders repository's tags.
 */
export const RepositoryTagTab: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ repo }) => (
    <div className="repository-graph-area">
        <div className="container">
            <div className="container-inner">
                <RepositoryReleasesTagsPage repo={repo} />
            </div>
        </div>
    </div>
)
