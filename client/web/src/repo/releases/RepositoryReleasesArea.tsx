import { FC } from 'react'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepositoryFields } from '../../graphql-operations'
import { RepoContainerContext } from '../RepoContainer'

import { RepositoryReleasesTagsPage } from './RepositoryReleasesTagsPage'

interface Props extends Pick<RepoContainerContext, 'repo'>, BreadcrumbSetters {
    repo: RepositoryFields
}

/**
 * Properties passed to all page components in the repository branches area.
 */
export interface RepositoryReleasesAreaPageProps {
    /**
     * The active repository.
     */
    repo: RepositoryFields
}

const BREADCRUMB = { key: 'tags', element: 'Tags' }

/**
 * Renders pages related to repository branches.
 */
export const RepositoryReleasesArea: FC<Props> = props => {
    const { useBreadcrumb, repo } = props

    useBreadcrumb(BREADCRUMB)

    return (
        <div className="repository-graph-area">
            <div className="container">
                <div className="container-inner">
                    <RepositoryReleasesTagsPage repo={repo} />
                </div>
            </div>
        </div>
    )
}
