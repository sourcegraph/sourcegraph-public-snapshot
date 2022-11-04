import React, { useMemo } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../../auth'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { HeroPage } from '../../../components/HeroPage'
import { RepositoryFields } from '../../../graphql-operations'
import { RouteDescriptor } from '../../../util/contributions'
import { CodeIntelConfigurationPageProps } from '../configuration/pages/CodeIntelConfigurationPage'
import { CodeIntelConfigurationPolicyPageProps } from '../configuration/pages/CodeIntelConfigurationPolicyPage'
import { CodeIntelRepositoryIndexConfigurationPageProps } from '../configuration/pages/CodeIntelRepositoryIndexConfigurationPage'
import { CodeIntelIndexesPageProps } from '../indexes/pages/CodeIntelIndexesPage'
import { CodeIntelIndexPageProps } from '../indexes/pages/CodeIntelIndexPage'
import { CodeIntelUploadPageProps } from '../uploads/pages/CodeIntelUploadPage'
import { CodeIntelUploadsPageProps } from '../uploads/pages/CodeIntelUploadsPage'

import { CodeIntelSidebar, CodeIntelSideBarGroups } from './CodeIntelSidebar'

export interface CodeIntelAreaRouteContext extends ThemeProps, TelemetryProps {
    repo: { id: string }
    authenticatedUser: AuthenticatedUser | null
}

export interface CodeIntelAreaRoute extends RouteDescriptor<CodeIntelAreaRouteContext> {}

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

export const routes: readonly CodeIntelAreaRoute[] = [
    {
        path: '/',
        exact: true,
        render: () => <Redirect to="./code-graph/uploads" />,
    },
    {
        path: '/uploads',
        exact: true,
        render: props => <CodeIntelUploadsPage {...props} />,
    },
    {
        path: '/uploads/:id',
        exact: true,
        render: props => <CodeIntelUploadPage {...props} />,
    },
    {
        path: '/indexes',
        exact: true,
        render: props => <CodeIntelIndexesPage {...props} />,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    },
    {
        path: '/indexes/:id',
        exact: true,
        render: props => <CodeIntelIndexPage {...props} />,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    },
    {
        path: '/configuration',
        exact: true,
        render: props => <CodeIntelConfigurationPage {...props} />,
    },
    {
        path: '/index-configuration',
        exact: true,
        render: props => <RepositoryIndexConfigurationPage {...props} />,
    },
    {
        path: '/configuration/:id',
        exact: true,
        render: props => <CodeIntelConfigurationPolicyPage {...props} />,
    },
]

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository page was not found."
    />
)

/**
 * Properties passed to all page components in the repository code navigation area.
 */
export interface RepositoryCodeIntelAreaPageProps
    extends ThemeProps,
        RouteComponentProps<{}>,
        BreadcrumbSetters,
        TelemetryProps {
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

/**
 * Renders pages related to repository code graph.
 */
export const RepositoryCodeIntelArea: React.FunctionComponent<
    React.PropsWithChildren<RepositoryCodeIntelAreaPageProps>
> = ({ match, useBreadcrumb, ...props }) => {
    useBreadcrumb(useMemo(() => ({ key: 'code-intelligence', element: 'Code graph data' }), []))

    return (
        <div className="container d-flex mt-3">
            <CodeIntelSidebar className="flex-0 mr-3" codeIntelSidebarGroups={sidebarRoutes} match={match} {...props} />

            <div className="flex-bounded">
                <Switch>
                    {routes.map(
                        ({ path, render, exact, condition = () => true }) =>
                            condition(props) && (
                                <Route
                                    path={match.url + path}
                                    exact={exact}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    render={routeComponentProps => render({ ...props, ...routeComponentProps })}
                                />
                            )
                    )}

                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        </div>
    )
}
