import { FC } from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { RepositoryFields } from '../../../graphql-operations'

import { BatchChangeRepoPage } from './BatchChangeRepoPage'

/**
 * Properties passed to all page components in the repository batch changes area.
 */
export interface RepositoryBatchChangesAreaPageProps extends ThemeProps, BreadcrumbSetters {
    /**
     * The active repository.
     */
    repo: RepositoryFields
}

const BREADCRUMB = { key: 'batch-changes', element: 'Batch Changes' }

/**
 * Renders pages related to repository batch changes.
 */
export const RepositoryBatchChangesArea: FC<RepositoryBatchChangesAreaPageProps> = props => {
    const { useBreadcrumb, repo, isLightTheme } = props

    useBreadcrumb(BREADCRUMB)

    return (
        <div className="repository-batch-changes-area container mt-3">
            <BatchChangeRepoPage repo={repo} isLightTheme={isLightTheme} />
        </div>
    )
}
