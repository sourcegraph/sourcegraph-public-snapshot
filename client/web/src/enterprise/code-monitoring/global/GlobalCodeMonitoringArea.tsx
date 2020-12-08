import React, { useMemo } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { CodeMonitoringProps } from '..'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { ThemeProps } from '../../../../../shared/src/theme'
import { AuthenticatedUser } from '../../../auth'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { BreadcrumbsProps, BreadcrumbSetters, Breadcrumbs } from '../../../components/Breadcrumbs'
import { lazyComponent } from '../../../util/lazyComponent'
import { CodeMonitoringPageProps } from '../CodeMonitoringPage'
import { CreateCodeMonitorPageProps } from '../CreateCodeMonitorPage'
import { ManageCodeMonitorPageProps } from '../ManageCodeMonitorPage'

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        CodeMonitoringProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

const CodeMonitoringPage = lazyComponent<CodeMonitoringPageProps, 'CodeMonitoringPage'>(
    () => import('../CodeMonitoringPage'),
    'CodeMonitoringPage'
)
const CreateCodeMonitorPage = lazyComponent<CreateCodeMonitorPageProps, 'CreateCodeMonitorPage'>(
    () => import('../CreateCodeMonitorPage'),
    'CreateCodeMonitorPage'
)

const ManageCodeMonitorPage = lazyComponent<ManageCodeMonitorPageProps, 'ManageCodeMonitorPage'>(
    () => import('../ManageCodeMonitorPage'),
    'ManageCodeMonitorPage'
)

/**
 * The global code monitoring area.
 */
export const GlobalCodeMonitoringArea: React.FunctionComponent<Props> = props => (
    <AuthenticatedCodeMonitoringArea {...props} />
)

interface AuthenticatedProps extends Props {
    authenticatedUser: AuthenticatedUser
}

export const AuthenticatedCodeMonitoringArea = withAuthenticatedUser<AuthenticatedProps>(({ match, ...outerProps }) => {
    const breadcrumbSetters = outerProps.useBreadcrumb(
        useMemo(
            () => ({
                key: 'Code Monitoring',
                element: <>Code Monitoring</>,
            }),
            []
        )
    )

    return (
        <div className="w-100">
            <Breadcrumbs breadcrumbs={outerProps.breadcrumbs} location={outerProps.location} />
            <div className="container web-content">
                {/* eslint-disable react/jsx-no-bind */}
                <Switch>
                    <Route
                        render={props => <CodeMonitoringPage {...outerProps} {...props} {...breadcrumbSetters} />}
                        path={match.url}
                        exact={true}
                    />
                    <Route
                        path={`${match.url}/new`}
                        render={props => <CreateCodeMonitorPage {...outerProps} {...props} {...breadcrumbSetters} />}
                        exact={true}
                    />
                    <Route
                        path={`${match.path}/:id`}
                        render={props => <ManageCodeMonitorPage {...outerProps} {...props} {...breadcrumbSetters} />}
                        exact={true}
                    />
                </Switch>
            </div>
        </div>
    )
})
