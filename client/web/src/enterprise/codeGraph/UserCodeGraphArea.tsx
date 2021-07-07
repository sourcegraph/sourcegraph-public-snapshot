import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { UserCodeGraphOverviewPage } from './UserCodeGraphOverviewPage'

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps {
    namespaceID: Scalars['ID']
}

export const UserCodeGraphArea: React.FunctionComponent<Props> = ({ match, namespaceID, ...outerProps }) => (
    <div className="pb-3">
        <Switch>
            <Route path={match.url} exact={true}>
                <UserCodeGraphOverviewPage {...outerProps} namespaceID={namespaceID} />
            </Route>
        </Switch>
    </div>
)
