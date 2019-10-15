import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { GlobalCampaignListPage } from './list/GlobalCampaignListPage'
import { CampaignDetails } from '../detail/CampaignDetails'
import { IUser } from '../../../../../shared/src/graphql/schema'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'

interface Props extends RouteComponentProps<{}> {
    authenticatedUser: IUser
}

/**
 * The global campaigns area.
 */
export const GlobalCampaignsArea = withAuthenticatedUser<Props>(({ match, ...layoutProps }) => (
    <div className="container mt-4">
        {/* eslint-disable react/jsx-no-bind */}
        <Switch>
            <Route render={props => <GlobalCampaignListPage {...props} />} path={match.url} exact={true} />
            <Route
                path={`${match.url}/new`}
                render={props => <CampaignDetails {...layoutProps} {...props} />}
                exact={true}
            />
            <Route
                path={`${match.url}/:campaignID`}
                render={({ match, ...props }: RouteComponentProps<{ campaignID: string }>) => (
                    <CampaignDetails {...layoutProps} {...props} campaignID={match.params.campaignID} />
                )}
            />
        </Switch>
        {/* eslint-enable react/jsx-no-bind */}
    </div>
))
export default GlobalCampaignsArea
