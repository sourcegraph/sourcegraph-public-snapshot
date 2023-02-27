import React from 'react'

import { Navigate } from 'react-router-dom'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { RepositoryFields } from '../../graphql-operations'

/**
 * Properties passed to all page components in the repository code navigation area.
 */
export interface RepositoryOwnAreaPageProps extends BreadcrumbSetters {
    /** The active repository. */
    repo: RepositoryFields
    authenticatedUser: AuthenticatedUser | null
}
const BREADCRUMB = { key: 'own', element: 'Ownership' }

export const RepositoryOwnPage: React.FunctionComponent<RepositoryOwnAreaPageProps> = ({
    useBreadcrumb,
    repo,
    authenticatedUser,
}) => {
    useBreadcrumb(BREADCRUMB)

    const [ownEnabled, status] = useFeatureFlag('search-ownership')

    if (status === 'initial') {
        return (
            <div className="container d-flex justify-content-center mt-3">
                <LoadingSpinner />
            </div>
        )
    }

    if (!ownEnabled) {
        return <Navigate to={repo.url} replace={true} />
    }

    return (
        <div className="container d-flex mt-3">
            <div className="flex-bounded">Hello world</div>
        </div>
    )
}
