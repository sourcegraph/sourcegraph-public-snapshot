import BrainIcon from 'mdi-react/BrainIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo } from 'react'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { HeroPage } from '@sourcegraph/web/src/components/HeroPage'

import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { RepositoryFields } from '../../../graphql-operations'
import { RouteDescriptor } from '../../../util/contributions'
import { lazyComponent } from '../../../util/lazyComponent'
import { CodeIntelDataRetentionConfigurationPageProps } from '../configuration/CodeIntelDataRetentionConfigurationPage'
import { CodeIntelIndexConfigurationPageProps } from '../configuration/CodeIntelIndexConfigurationPage'
import { CodeIntelIndexSchedulerConfigurationPageProps } from '../configuration/CodeIntelIndexSchedulerConfiguration'
import { CodeIntelIndexPageProps } from '../detail/CodeIntelIndexPage'
import { CodeIntelUploadPageProps } from '../detail/CodeIntelUploadPage'
import { CodeIntelIndexesPageProps } from '../list/CodeIntelIndexesPage'
import { CodeIntelUploadsPageProps } from '../list/CodeIntelUploadsPage'

import { CodeIntelSidebar, CodeIntelSideBarGroups } from './CodeIntelSidebar'

export interface CodeIntelAreaRouteContext extends ThemeProps, TelemetryProps {
    repo: { id: string }
}

export interface CodeIntelAreaRoute extends RouteDescriptor<CodeIntelAreaRouteContext> {}

const CodeIntelUploadsPage = lazyComponent<CodeIntelUploadsPageProps, 'CodeIntelUploadsPage'>(
    () => import('../../codeintel/list/CodeIntelUploadsPage'),
    'CodeIntelUploadsPage'
)
const CodeIntelUploadPage = lazyComponent<CodeIntelUploadPageProps, 'CodeIntelUploadPage'>(
    () => import('../../codeintel/detail/CodeIntelUploadPage'),
    'CodeIntelUploadPage'
)

const CodeIntelDataRetentionConfigurationPage = lazyComponent<
    CodeIntelDataRetentionConfigurationPageProps,
    'CodeIntelDataRetentionConfigurationPage'
>(
    () => import('../../codeintel/configuration/CodeIntelDataRetentionConfigurationPage'),
    'CodeIntelDataRetentionConfigurationPage'
)

const CodeIntelIndexesPage = lazyComponent<CodeIntelIndexesPageProps, 'CodeIntelIndexesPage'>(
    () => import('../../codeintel/list/CodeIntelIndexesPage'),
    'CodeIntelIndexesPage'
)
const CodeIntelIndexPage = lazyComponent<CodeIntelIndexPageProps, 'CodeIntelIndexPage'>(
    () => import('../../codeintel/detail/CodeIntelIndexPage'),
    'CodeIntelIndexPage'
)

const CodeIntelIndexConfigurationPage = lazyComponent<
    CodeIntelIndexConfigurationPageProps,
    'CodeIntelIndexConfigurationPage'
>(() => import('../../codeintel/configuration/CodeIntelIndexConfigurationPage'), 'CodeIntelIndexConfigurationPage')

const CodeIntelIndexScheduleConfigurationPage = lazyComponent<
    CodeIntelIndexSchedulerConfigurationPageProps,
    'CodeIntelIndexScheduleConfigurationPage'
>(
    () => import('../../codeintel/configuration/CodeIntelIndexScheduleConfigurationPage'),
    'CodeIntelIndexScheduleConfigurationPage'
)

export const routes: readonly CodeIntelAreaRoute[] = [
    {
        path: '/',
        exact: true,
        render: () => <Redirect to="./code-intelligence/uploads" />,
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
        path: '/data-retention-configuration',
        exact: true,
        render: props => <CodeIntelDataRetentionConfigurationPage {...props} />,
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
        path: '/index-configuration',
        exact: true,
        render: props => <CodeIntelIndexConfigurationPage {...props} />,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    },
    {
        path: '/index-schedule-configuration',
        exact: true,
        render: props => <CodeIntelIndexScheduleConfigurationPage {...props} />,
    },
]

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository page was not found."
    />
)

/**
 * Properties passed to all page components in the repository code intelligence area.
 */
export interface RepositoryCodeIntelAreaPageProps
    extends ThemeProps,
        RouteComponentProps<{}>,
        BreadcrumbSetters,
        TelemetryProps {
    /** The active repository. */
    repo: RepositoryFields
}

const sidebarRoutes: CodeIntelSideBarGroups = [
    {
        header: { label: 'Precise intelligence', icon: BrainIcon },
        items: [
            {
                to: '/uploads',
                label: 'Uploads',
            },
            {
                to: '/data-retention-configuration',
                label: 'Data retention configuration',
            },
            {
                to: '/indexes',
                label: 'Auto indexing',
                condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
            },
            {
                to: '/index-configuration',
                label: 'Index configuration',
                condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
            },
            {
                to: '/index-schedule-configuration',
                label: 'Index schedule configuration',
                condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
            },
        ],
    },
]

/**
 * Renders pages related to repository code intelligence.
 */
export const RepositoryCodeIntelArea: React.FunctionComponent<RepositoryCodeIntelAreaPageProps> = ({
    match,
    useBreadcrumb,
    ...props
}) => {
    useBreadcrumb(useMemo(() => ({ key: 'code-intelligence', element: 'Code Intelligence' }), []))

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
