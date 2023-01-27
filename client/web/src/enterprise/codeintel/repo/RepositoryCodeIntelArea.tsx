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
import { CodeIntelInferenceConfigurationPageProps } from '../configuration/pages/CodeIntelInferenceConfigurationPage'
import { CodeIntelRepositoryIndexConfigurationPageProps } from '../configuration/pages/CodeIntelRepositoryIndexConfigurationPage'

import { CodeIntelPreciseIndexesPageProps } from '../indexes/pages/CodeIntelPreciseIndexesPage'
import { CodeIntelPreciseIndexPageProps } from '../indexes/pages/CodeIntelPreciseIndexPage'
import { CodeIntelSidebar, CodeIntelSideBarGroups } from './CodeIntelSidebar'

export interface CodeIntelAreaRouteContext extends ThemeProps, TelemetryProps {
    repo: { id: string; name: string }
    authenticatedUser: AuthenticatedUser | null
}

export interface CodeIntelAreaRoute extends RouteDescriptor<CodeIntelAreaRouteContext> {}

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

export const routes: readonly CodeIntelAreaRoute[] = [
    {
        path: '/',
        exact: true,
        render: () => <Redirect to="./code-graph/indexes" />,
    },
    {
        path: '/indexes',
        exact: true,
        render: props => <CodeIntelPreciseIndexesPage {...props} />,
    },
    {
        path: '/indexes/:id',
        exact: true,
        render: props => <CodeIntelPreciseIndexPage {...props} />,
    },
    // {
    //     path: '/uploads',
    //     exact: true,
    //     render: props => <CodeIntelUploadsPage {...props} />,
    // },
    // {
    //     path: '/uploads/:id',
    //     exact: true,
    //     render: props => <CodeIntelUploadPage {...props} />,
    // },
    // {
    //     path: '/indexes',
    //     exact: true,
    //     render: props => <CodeIntelIndexesPage {...props} />,
    //     condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    // },
    // {
    //     path: '/indexes/:id',
    //     exact: true,
    //     render: props => <CodeIntelIndexPage {...props} />,
    //     condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    // },
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
        path: '/inference-configuration',
        exact: true,
        render: props => <CodeIntelInferenceConfigurationPage {...props} />,
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
