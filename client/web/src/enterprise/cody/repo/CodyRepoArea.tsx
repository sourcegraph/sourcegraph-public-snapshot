import { FC } from 'react'

import { Navigate, Route, Routes } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../../auth'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { NotFoundPage } from '../../../components/HeroPage'
import { RepositoryFields } from '../../../graphql-operations'
import { RouteV6Descriptor } from '../../../util/contributions'
import { CodeIntelConfigurationPolicyPageProps } from '../../codeintel/configuration/pages/CodeIntelConfigurationPolicyPage'
import { CodyConfigurationPage } from '../configuration/pages/CodyConfigurationPage'

import { CodyRepoSidebar, CodyRepoSidebarGroups } from './CodyRepoSidebar'

export interface CodyRepoAreaRouteContext extends TelemetryProps {
    repo: { id: string; name: string }
    authenticatedUser: AuthenticatedUser | null
}

export interface CodyRepoAreaRoute extends RouteV6Descriptor<CodyRepoAreaRouteContext> {}

const CodeIntelConfigurationPolicyPage = lazyComponent<
    CodeIntelConfigurationPolicyPageProps,
    'CodeIntelConfigurationPolicyPage'
>(
    () => import('../../codeintel/configuration/pages/CodeIntelConfigurationPolicyPage'),
    'CodeIntelConfigurationPolicyPage'
)

export const codyRepoAreaRoutes: readonly CodyRepoAreaRoute[] = [
    {
        path: '/',
        render: () => <Navigate to="./configuration" replace={true} />,
    },
    {
        path: '/configuration',
        render: props => <CodyConfigurationPage {...props} />,
    },
    {
        path: '/configuration/:id',
        render: props => <CodeIntelConfigurationPolicyPage {...props} domain="embeddings" />,
    },
]

export interface CodyRepoAreaProps extends BreadcrumbSetters, TelemetryProps {
    repo: RepositoryFields
    authenticatedUser: AuthenticatedUser | null
}

const sidebarRoutes: CodyRepoSidebarGroups = [
    {
        header: { label: 'Embeddings' },
        items: [
            {
                to: '/configuration',
                label: 'Configuration policies',
            },
        ],
    },
]

const BREADCRUMB = { key: 'embeddings', element: 'Embeddings' }

export const CodyRepoArea: FC<CodyRepoAreaProps> = props => {
    const { useBreadcrumb, repo } = props

    useBreadcrumb(BREADCRUMB)

    return (
        <div className="container d-flex mt-3">
            <CodyRepoSidebar className="flex-0 mr-3" codyRepoSidebarGroups={sidebarRoutes} repo={repo} />

            <div className="flex-bounded">
                <Routes>
                    {codyRepoAreaRoutes.map(
                        ({ path, render, condition = () => true }) =>
                            condition(props) && (
                                <Route
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    path={path}
                                    element={render(props)}
                                />
                            )
                    )}

                    <Route path="*" element={<NotFoundPage pageType="repository" />} />
                </Routes>
            </div>
        </div>
    )
}
