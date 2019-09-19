import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { GlobalCampaignListPage } from './list/GlobalCampaignListPage'
import { NewCampaignPage } from '../new/NewCampaignPage'
import { CampaignDetails } from '../detail/CampaignDetails'

interface Props extends RouteComponentProps<{}> {}

/**
 * The global campaigns area.
 */
export const GlobalCampaignsArea: React.FunctionComponent<Props> = ({ match }) => (
    <div className="container mt-4">
        {/* eslint-disable react/jsx-no-bind */}
        <Switch>
            <Route render={() => <GlobalCampaignListPage />} path={match.url} exact={true} />
            <Route path={`${match.url}/new`} render={() => <NewCampaignPage />} exact={true} />
            <Route
                path={`${match.url}/:campaignID`}
                render={(routeComponentProps: RouteComponentProps<{ campaignID: string }>) => (
                    <CampaignDetails campaignID={routeComponentProps.match.params.campaignID} />
                )}
            />
        </Switch>
        {/* eslint-enable react/jsx-no-bind */}
    </div>
)
export default GlobalCampaignsArea
