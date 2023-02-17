import { FC } from 'react'

import { Routes, Route, Navigate } from 'react-router-dom-v5-compat'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../../auth'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { NotFoundPage } from '../../../components/HeroPage'
import { RedirectRoute } from '../../../components/RedirectRoute'
import { RepositoryFields } from '../../../graphql-operations'
import { RouteV6Descriptor } from '../../../util/contributions'
import { CodeIntelConfigurationPageProps } from '../configuration/pages/CodeIntelConfigurationPage'
import { CodeIntelConfigurationPolicyPageProps } from '../configuration/pages/CodeIntelConfigurationPolicyPage'
import { CodeIntelInferenceConfigurationPageProps } from '../configuration/pages/CodeIntelInferenceConfigurationPage'
import { CodeIntelRepositoryIndexConfigurationPageProps } from '../configuration/pages/CodeIntelRepositoryIndexConfigurationPage'
import { RepoDashboardPageProps } from '../dashboard/pages/RepoDashboardPage'
import { CodeIntelPreciseIndexesPageProps } from '../indexes/pages/CodeIntelPreciseIndexesPage'
import { CodeIntelPreciseIndexPageProps } from '../indexes/pages/CodeIntelPreciseIndexPage'

import { CodeIntelSidebar, CodeIntelSideBarGroups } from './CodeIntelSidebar'

export interface CodeIntelAreaRouteContext extends ThemeProps, TelemetryProps {
    repo: { id: string; name: string }
    authenticatedUser: AuthenticatedUser | null
}

export interface CodeIntelAreaRoute extends RouteV6Descriptor<CodeIntelAreaRouteContext> {}

const RepoDashboardPage = lazyComponent<RepoDashboardPageProps, 'RepoDashboardPage'>(
    () => import('../dashboard/pages/RepoDashboardPage'),
    'RepoDashboardPage'
)

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

const RepositoryIndexConfigurationPage = lazyComponent<
    CodeIntelRepositoryIndexConfigurationPageProps,
    'CodeIntelRepositoryIndexConfigurationPage'
>(
    () => import('../configuration/pages/CodeIntelRepositoryIndexConfigurationPage'),
    'CodeIntelRepositoryIndexConfigurationPage'
)

const CodeIntelConfigurationPolicyPage = lazyComponent<
    CodeIntelConfigurationPolicyPageProps,
    'CodeIntelConfigurationPolicyPage'
>(() => import('../configuration/pages/CodeIntelConfigurationPolicyPage'), 'CodeIntelConfigurationPolicyPage')

export const codeIntelAreaRoutes: readonly CodeIntelAreaRoute[] = [
    {
        path: '/',
        render: () => <Navigate to="./dashboard" replace={true} />,
    },
    {
        path: '/dashboard',
        render: props => <RepoDashboardPage {...props} />,
    },
    {
        path: '/indexes',
        render: props => <CodeIntelPreciseIndexesPage {...props} />,
    },
    {
        path: '/indexes/:id',
        render: props => <CodeIntelPreciseIndexPage {...props} />,
    },
    {
        path: '/uploads/:id',
        render: () => (
            <RedirectRoute
                getRedirectURL={({ params }) =>
                    `../indexes/${btoa(`PreciseIndex:"U:${(atob(params.id!).match(/(\d+)/) ?? [''])[0]}"`)}`
                }
            />
        ),
    },
    {
        path: '/configuration',
        render: props => <CodeIntelConfigurationPage {...props} />,
    },
    {
        path: '/index-configuration',
        render: props => <RepositoryIndexConfigurationPage {...props} />,
    },
    {
        path: '/inference-configuration',
        render: props => <CodeIntelInferenceConfigurationPage {...props} />,
    },
    {
        path: '/configuration/:id',
        render: props => <CodeIntelConfigurationPolicyPage {...props} />,
    },
]

/**
 * Properties passed to all page components in the repository code navigation area.
 */
export interface RepositoryCodeIntelAreaPageProps extends ThemeProps, BreadcrumbSetters, TelemetryProps {
    /** The active repository. */
    repo: RepositoryFields
    authenticatedUser: AuthenticatedUser | null
}

const sidebarRoutes: CodeIntelSideBarGroups = [
    {
        header: { label: 'Code graph data' },
        items: [
            {
                to: '/dashboard',
                label: 'Dashboard',
            },
            {
                to: '/indexes',
                label: 'Precise indexes',
            },
            {
                to: '/configuration',
                label: 'Configuration policies',
            },
            {
                to: '/index-configuration',
                label: 'Auto-index configuration',
                condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
            },
        ],
    },
]

const BREADCRUMB = { key: 'code-intelligence', element: 'Code graph data' }

/**
 * Renders pages related to repository code graph.
 */
export const RepositoryCodeIntelArea: FC<RepositoryCodeIntelAreaPageProps> = props => {
    const { useBreadcrumb } = props

    useBreadcrumb(BREADCRUMB)

    return (
        <div className="container d-flex mt-3">
            <CodeIntelSidebar className="flex-0 mr-3" codeIntelSidebarGroups={sidebarRoutes} {...props} />

            <div className="flex-bounded">
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

                    <Route element={<NotFoundPage pageType="repository" />} />
                </Routes>
            </div>
        </div>
    )
}
