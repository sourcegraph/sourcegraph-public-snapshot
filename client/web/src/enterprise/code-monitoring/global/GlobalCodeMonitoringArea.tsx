import React from 'react'

import { RouteComponentProps, Switch } from 'react-router'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { H2, Text } from '@sourcegraph/wildcard'

import { XCompatRoute } from '../../../XCompatRoute'
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
}) => {
    console.log('GlobalCodeMonitoringArea', match)

    return (
        <div className="w-100">
            <Page>
                <H2>GlobalCodeMonitoringArea content</H2>
                <Text className="mb-5">`match.url` value is: {match.url}</Text>
                <Switch>
                    <XCompatRoute
                        path="/code-monitoring/new"
                        pathV6="new"
                        render={(props: RouteComponentProps<{}>) => (
                            <CreateCodeMonitorPage {...outerProps} {...props} />
                        )}
                    />
                    <XCompatRoute
                        path="/code-monitoring/:id"
                        pathV6=":id"
                        render={(props: RouteComponentProps<{ id: string }>) => {
                            console.log('render route ManageCodeMonitorPage')

                            return <ManageCodeMonitorPage {...outerProps} {...props} />
                        }}
                        exact={true}
                    />
                    <XCompatRoute
                        path="/code-monitoring"
                        pathV6=""
                        render={(props: RouteComponentProps<{}>) => <CodeMonitoringPage {...outerProps} {...props} />}
                    />
                </Switch>
            </Page>
        </div>
    )
}
