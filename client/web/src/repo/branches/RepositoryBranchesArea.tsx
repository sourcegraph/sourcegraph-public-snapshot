import type { FC } from 'react'

import { Routes, Route } from 'react-router-dom'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import type { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { NotFoundPage } from '../../components/HeroPage'
import type { RepositoryFields } from '../../graphql-operations'

import { RepositoryBranchesAllPage } from './RepositoryBranchesAllPage'
import { RepositoryBranchesNavbar } from './RepositoryBranchesNavbar'
import { RepositoryBranchesOverviewPage } from './RepositoryBranchesOverviewPage'

interface Props extends BreadcrumbSetters, TelemetryV2Props {
    repo: RepositoryFields
}

/**
 * Properties passed to all page components in the repository branches area.
 */
export interface RepositoryBranchesAreaPageProps extends TelemetryV2Props {
    /**
     * The active repository.
     */
    repo: RepositoryFields
}

const BREADCRUMB = { key: 'branches', element: 'Branches' }

/**
 * Renders pages related to repository branches.
 */
export const RepositoryBranchesArea: FC<Props> = props => {
    const { useBreadcrumb, repo, telemetryRecorder } = props

    useBreadcrumb(BREADCRUMB)

    return (
        <div className="repository-branches-area container px-3">
            <RepositoryBranchesNavbar className="my-3" repo={repo.name} />
            <Routes>
                <Route
                    path="all"
                    element={<RepositoryBranchesAllPage repo={repo} telemetryRecorder={telemetryRecorder} />}
                />
                <Route
                    path=""
                    element={<RepositoryBranchesOverviewPage repo={repo} telemetryRecorder={telemetryRecorder} />}
                />
                <Route path="*" element={<NotFoundPage pageType="repository branches" />} />
            </Routes>
        </div>
    )
}
