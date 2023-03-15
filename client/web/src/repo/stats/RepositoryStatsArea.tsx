import { FC } from 'react'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepositoryFields } from '../../graphql-operations'

import { RepositoryStatsContributorsPage } from './RepositoryStatsContributorsPage'

interface Props extends BreadcrumbSetters {
    repo: RepositoryFields | undefined
    globbing: boolean
}

/**
 * Properties passed to all page components in the repository stats area.
 */
export interface RepositoryStatsAreaPageProps {
    /**
     * The active repository.
     */
    repo: RepositoryFields
}

const BREADCRUMB = { key: 'contributors', element: 'Contributors' }

/**
 * Renders pages related to repository stats.
 */
export const RepositoryStatsArea: FC<Props> = props => {
    const { useBreadcrumb, repo, globbing } = props
    useBreadcrumb(BREADCRUMB)

    return (
        <div className="repository-stats-area container mt-3">
            {repo ? <RepositoryStatsContributorsPage repo={repo} globbing={globbing} /> : <LoadingSpinner />}
        </div>
    )
}
