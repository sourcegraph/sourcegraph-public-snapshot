import { FC } from 'react'

import { Redirect } from 'react-router'
import { Routes, Route } from 'react-router-dom-v5-compat'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../../auth'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { NotFoundPage } from '../../../components/HeroPage'
import { RepositoryFields } from '../../../graphql-operations'
import { RouteV6Descriptor } from '../../../util/contributions'
import { CodeIntelConfigurationPageProps } from '../configuration/pages/CodeIntelConfigurationPage'
import { CodeIntelConfigurationPolicyPageProps } from '../configuration/pages/CodeIntelConfigurationPolicyPage'
import { CodeIntelInferenceConfigurationPageProps } from '../configuration/pages/CodeIntelInferenceConfigurationPage'
import { CodeIntelRepositoryIndexConfigurationPageProps } from '../configuration/pages/CodeIntelRepositoryIndexConfigurationPage'
import { CodeIntelIndexesPageProps } from '../indexes/pages/CodeIntelIndexesPage'
import { CodeIntelIndexPageProps } from '../indexes/pages/CodeIntelIndexPage'
import { CodeIntelUploadPageProps } from '../uploads/pages/CodeIntelUploadPage'
import { CodeIntelUploadsPageProps } from '../uploads/pages/CodeIntelUploadsPage'

import { CodeIntelSidebar, CodeIntelSideBarGroups } from './CodeIntelSidebar'

export interface CodeIntelAreaRouteContext extends ThemeProps, TelemetryProps {
    repo: { id: string; name: string }
    authenticatedUser: AuthenticatedUser | null
}

export interface CodeIntelAreaRoute extends RouteV6Descriptor<CodeIntelAreaRouteContext> {}

const CodeIntelUploadsPage = lazyComponent<CodeIntelUploadsPageProps, 'CodeIntelUploadsPage'>(
    () => import('../uploads/pages/CodeIntelUploadsPage'),
    'CodeIntelUploadsPage'
)
const CodeIntelUploadPage = lazyComponent<CodeIntelUploadPageProps, 'CodeIntelUploadPage'>(
    () => import('../uploads/pages/CodeIntelUploadPage'),
    'CodeIntelUploadPage'
)

const CodeIntelIndexesPage = lazyComponent<CodeIntelIndexesPageProps, 'CodeIntelIndexesPage'>(
    () => import('../indexes/pages/CodeIntelIndexesPage'),
    'CodeIntelIndexesPage'
)
const CodeIntelIndexPage = lazyComponent<CodeIntelIndexPageProps, 'CodeIntelIndexPage'>(
    () => import('../indexes/pages/CodeIntelIndexPage'),
    'CodeIntelIndexPage'
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
        render: () => <Redirect to="./code-graph/uploads" />,
    },
    {
        path: '/uploads',
        render: props => <CodeIntelUploadsPage {...props} />,
    },
    {
        path: '/uploads/:id',
        render: props => <CodeIntelUploadPage {...props} />,
    },
    {
        path: '/indexes',
        render: props => <CodeIntelIndexesPage {...props} />,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    },
    {
        path: '/indexes/:id',
        render: props => <CodeIntelIndexPage {...props} />,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
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
                to: '/uploads',
                label: 'Uploads',
            },
            {
                to: '/indexes',
                label: 'Auto-indexing',
                condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
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

                    <Route element={<NotFoundPage />} />
                </Routes>
            </div>
        </div>
    )
}
