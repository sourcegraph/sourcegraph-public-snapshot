import React from 'react'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import { CodeMonitoringProps } from '..'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { ThemeProps } from '../../../../../shared/src/theme'
import { AuthenticatedUser } from '../../../auth'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
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
        CodeMonitoringProps,
        SettingsCascadeProps {
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
    if (!outerProps.authenticatedUser) {
        return <Redirect to="/sign-in" />
    }

    return (
        <div className="w-100">
            <div className="container web-content">
                {/* eslint-disable react/jsx-no-bind */}
                <Switch>
                    <Route
                        render={props => <CodeMonitoringPage {...outerProps} {...props} />}
                        path={match.url}
                        exact={true}
                    />
                    <Route
                        path={`${match.url}/new`}
                        render={props => <CreateCodeMonitorPage {...outerProps} {...props} />}
                        exact={true}
                    />
                    <Route
                        path={`${match.path}/:id`}
                        render={props => <ManageCodeMonitorPage {...outerProps} {...props} />}
                        exact={true}
                    />
                </Switch>
            </div>
        </div>
    )
})
