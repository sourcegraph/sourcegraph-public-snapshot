import React from 'react'

import { Route, RouteComponentProps, Switch } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../../auth'
import { Page } from '../../../components/Page'

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps,
        SettingsCascadeProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

const CodeMonitoringPage = lazyComponent(() => import('../CodeMonitoringPage'), 'CodeMonitoringPage')
const CreateCodeMonitorPage = lazyComponent(() => import('../CreateCodeMonitorPage'), 'CreateCodeMonitorPage')
const ManageCodeMonitorPage = lazyComponent(() => import('../ManageCodeMonitorPage'), 'ManageCodeMonitorPage')

/**
 * The global code monitoring area.
 */
export const GlobalCodeMonitoringArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    match,
    ...outerProps
}) => (
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
