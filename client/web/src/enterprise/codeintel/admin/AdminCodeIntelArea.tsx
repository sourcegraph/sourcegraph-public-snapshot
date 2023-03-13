import { FC } from 'react'

import { Routes, Route, Navigate, useParams } from 'react-router-dom'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { NotFoundPage } from '../../../components/HeroPage'
import { SiteAdminAreaRouteContext } from '../../../site-admin/SiteAdminArea'
import { RouteV6Descriptor } from '../../../util/contributions'
import { CodeIntelConfigurationPageProps } from '../configuration/pages/CodeIntelConfigurationPage'
import { CodeIntelConfigurationPolicyPageProps } from '../configuration/pages/CodeIntelConfigurationPolicyPage'
import { CodeIntelInferenceConfigurationPageProps } from '../configuration/pages/CodeIntelInferenceConfigurationPage'
import { GlobalDashboardPageProps } from '../dashboard/pages/GlobalDashboardPage'
import { CodeIntelPreciseIndexesPageProps } from '../indexes/pages/CodeIntelPreciseIndexesPage'
import { CodeIntelPreciseIndexPageProps } from '../indexes/pages/CodeIntelPreciseIndexPage'

export interface AdminCodeIntelAreaRouteContext extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
}

export interface AdminCodeIntelAreaRoute extends RouteV6Descriptor<AdminCodeIntelAreaRouteContext> {}

const CodeIntelPreciseIndexesPage = lazyComponent<CodeIntelPreciseIndexesPageProps, 'CodeIntelPreciseIndexesPage'>(
    () => import('../indexes/pages/CodeIntelPreciseIndexesPage'),
    'CodeIntelPreciseIndexesPage'
)
const CodeIntelPreciseIndexPage = lazyComponent<CodeIntelPreciseIndexPageProps, 'CodeIntelPreciseIndexPage'>(
    () => import('../indexes/pages/CodeIntelPreciseIndexPage'),
    'CodeIntelPreciseIndexPage'
)

const CodeIntelConfigurationPage = lazyComponent<CodeIntelConfigurationPageProps, 'CodeIntelConfigurationPage'>(
    () => import('../configuration/pages/CodeIntelConfigurationPage'),
    'CodeIntelConfigurationPage'
)

const CodeIntelInferenceConfigurationPage = lazyComponent<
    CodeIntelInferenceConfigurationPageProps,
    'CodeIntelInferenceConfigurationPage'
>(() => import('../configuration/pages/CodeIntelInferenceConfigurationPage'), 'CodeIntelInferenceConfigurationPage')

const CodeIntelConfigurationPolicyPage = lazyComponent<
    CodeIntelConfigurationPolicyPageProps,
    'CodeIntelConfigurationPolicyPage'
>(() => import('../configuration/pages/CodeIntelConfigurationPolicyPage'), 'CodeIntelConfigurationPolicyPage')

const GlobalDashboardPage = lazyComponent<GlobalDashboardPageProps, 'GlobalDashboardPage'>(
    () => import('../dashboard/pages/GlobalDashboardPage'),
    'GlobalDashboardPage'
)

export const codeIntelAreaRoutes: readonly AdminCodeIntelAreaRoute[] = [
    // Code intelligence dashboard routes
    {
        exact: true,
        path: '/',
        render: () => <Navigate to="./code-graph/dashboard" replace={true} />,
    },
    {
        exact: true,
        path: '/dashboard',
        render: props => <GlobalDashboardPage {...props} />,
    },

    // Precise index routes
    {
        path: '/indexes',
        render: props => <CodeIntelPreciseIndexesPage {...props} />,
    },
    {
        path: '/indexes/:id',
        render: props => <CodeIntelPreciseIndexPage {...props} />,
    },

    // Code graph configuration
    {
        path: '/configuration',
        render: props => <CodeIntelConfigurationPage {...props} />,
    },
    {
        path: '/configuration/:id',
        render: props => <CodeIntelConfigurationPolicyPage {...props} />,
    },
    {
        path: '/inference-configuration',
        render: props => <CodeIntelInferenceConfigurationPage {...props} />,
    },

    // Legacy routes
    {
        path: '/uploads/:id',
        render: () => <NavigateToLegacyUploadPage />,
    },
].filter(Boolean) as readonly AdminCodeIntelAreaRoute[]

/**
 * Properties passed to all page components in the repository code navigation area.
 */
export interface AdminCodeIntelAreaPageProps extends SiteAdminAreaRouteContext {}

/**
 * Renders pages related to repository code graph.
 */
export const AdminCodeIntelArea: FC<AdminCodeIntelAreaPageProps> = props => (
    <Routes>
        {codeIntelAreaRoutes.map(
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
)

function NavigateToLegacyUploadPage(): JSX.Element {
    const { id = '' } = useParams<{ id: string }>()
    return (
        <Navigate
            to={`/site-admin/code-graph/indexes/${btoa(`PreciseIndex:"U:${(atob(id).match(/(\d+)/) ?? [''])[0]}"`)}`}
        />
    )
}
