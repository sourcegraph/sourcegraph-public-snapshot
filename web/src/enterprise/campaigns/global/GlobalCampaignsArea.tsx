import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { GlobalCampaignListPage } from './list/GlobalCampaignListPage'
import { CampaignDetails } from '../detail/CampaignDetails'
import { IUser } from '../../../../../shared/src/graphql/schema'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignUpdateSelection } from '../detail/CampaignUpdateSelection'

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps {
    authenticatedUser: IUser
    isSourcegraphDotCom: boolean
}

/**
 * The global campaigns area.
 */
export const GlobalCampaignsArea = withAuthenticatedUser<Props>(({ match, ...outerProps }) => (
    <div className="container mt-4">
        {/* eslint-disable react/jsx-no-bind */}
        <Switch>
            <Route
                render={props => <GlobalCampaignListPage {...outerProps} {...props} />}
                path={match.url}
                exact={true}
            />
            <Route
                path={`${match.url}/new`}
                render={props => <CampaignDetails {...outerProps} {...props} />}
                exact={true}
            />
            <Route
                path={`${match.url}/update`}
                render={props => <CampaignUpdateSelection {...outerProps} {...props} />}
                exact={true}
            />
            <Route
                path={`${match.url}/:campaignID`}
                render={({ match, ...props }: RouteComponentProps<{ campaignID: string }>) => (
                    <CampaignDetails {...outerProps} {...props} campaignID={match.params.campaignID} />
                )}
            />
        </Switch>
        {/* eslint-enable react/jsx-no-bind */}
    </div>
))
