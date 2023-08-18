import { type FC, useMemo } from 'react'

import type { BreadcrumbSetters } from '../../components/Breadcrumbs'
import type { RepositoryFields } from '../../graphql-operations'
import { isPackageServiceType } from '../packages/isPackageServiceType'
import type { RepoContainerContext } from '../RepoContainer'

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

const TAGS_BREADCRUMB = { key: 'tags', element: 'Tags' }
const VERSIONS_BREADCRUMB = { key: 'versions', element: 'Versions' }

/**
 * Renders pages related to repository branches.
 */
export const RepositoryReleasesArea: FC<Props> = props => {
    const { useBreadcrumb, repo } = props

    const isPackage = useMemo(
        () => isPackageServiceType(repo?.externalRepository.serviceType),
        [repo.externalRepository.serviceType]
    )

    useBreadcrumb(isPackage ? VERSIONS_BREADCRUMB : TAGS_BREADCRUMB)

    return (
        <div className="repository-graph-area">
            <div className="container">
                <div className="container-inner">
                    <RepositoryReleasesTagsPage repo={repo} isPackage={isPackage} />
                </div>
            </div>
        </div>
    )
}
