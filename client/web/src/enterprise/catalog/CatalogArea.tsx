import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { Switch, Route, useRouteMatch, RouteComponentProps } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { HeroPage } from '../../components/HeroPage'

import { GroupPage } from './pages/group/GroupPage'
import { GlobalOverviewPage } from './pages/overview/GlobalOverviewPage'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {}

/**
 * The main entrypoint to the catalog UI.
 */
export const CatalogArea: React.FunctionComponent<Props> = ({ telemetryService, ...props }) => {
    const match = useRouteMatch()

    return (
        <Switch>
            <Route path={`${match.url}/groups/:name`}>
                {(matchProps: RouteComponentProps<{ name: string }>) => (
                    <GroupPage
                        {...props}
                        groupName={matchProps.match.params.name}
                        telemetryService={telemetryService}
                    />
                )}
            </Route>
            <Route path={[match.url, `${match.url}/graph`]} exact={true}>
                <GlobalOverviewPage basePathname={match.url} telemetryService={telemetryService} />
            </Route>
            <Route>
                <HeroPage icon={MapSearchIcon} title="404: Not Found" />
            </Route>
        </Switch>
    )
}
