import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { SettingsCascadeProps } from '@sourcegraph/client-api'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../../auth'
import { Page } from '../../../components/Page'
import { FeatureFlagProps } from '../../../featureFlags/featureFlags'

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps,
        SettingsCascadeProps,
        FeatureFlagProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

const CodeMonitoringPage = lazyComponent(() => import('../CodeMonitoringPage'), 'CodeMonitoringPage')
const CreateCodeMonitorPage = lazyComponent(() => import('../CreateCodeMonitorPage'), 'CreateCodeMonitorPage')
const ManageCodeMonitorPage = lazyComponent(() => import('../ManageCodeMonitorPage'), 'ManageCodeMonitorPage')

/**
 * The global code monitoring area.
 */
export const GlobalCodeMonitoringArea: React.FunctionComponent<Props> = ({ match, ...outerProps }) => (
    <div className="w-100">
        <Page>
            <Switch>
                <Route
                    path={match.url}
                    render={props => <CodeMonitoringPage {...outerProps} {...props} />}
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
        </Page>
    </div>
)
