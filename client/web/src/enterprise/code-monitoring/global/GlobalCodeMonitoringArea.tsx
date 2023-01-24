import React from 'react'

import { RouteComponentProps, Switch } from 'react-router'
import { CompatRoute } from 'react-router-dom-v5-compat'

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
                <CompatRoute
                    path={match.url}
                    render={(props: RouteComponentProps<{}>) => <CodeMonitoringPage {...outerProps} {...props} />}
                    exact={true}
                />
                <CompatRoute
                    path={`${match.url}/new`}
                    render={(props: RouteComponentProps<{}>) => <CreateCodeMonitorPage {...outerProps} {...props} />}
                    exact={true}
                />
                <CompatRoute
                    path={`${match.path}/:id`}
                    render={(props: RouteComponentProps<{ id: string }>) => (
                        <ManageCodeMonitorPage {...outerProps} {...props} />
                    )}
                    exact={true}
                />
            </Switch>
        </Page>
    </div>
)
