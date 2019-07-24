import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { CampaignArea } from '../detail/CampaignArea'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { GlobalCampaignsListPage } from './list/GlobalCampaignsListPage'

interface Props
    extends RouteComponentProps<{}>,
        Pick<
            NamespaceCampaignsAreaContext,
            Exclude<keyof NamespaceCampaignsAreaContext, 'namespace' | 'setBreadcrumbItem'>
        > {}

/**
 * The global campaigns area.
 */
export const GlobalCampaignsArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: NamespaceCampaignsAreaContext = {
        ...props,
        campaignsURL: match.url,
    }
    return (
        <Switch>
            <Route path={context.campaignsURL} exact={true}>
                <div className="container mt-4">
                    <GlobalCampaignsListPage {...context} />
                </div>
            </Route>
            <Route
                path={`${context.campaignsURL}/:campaignID`}
                // tslint:disable-next-line: jsx-no-lambda
                render={(routeComponentProps: RouteComponentProps<{ campaignID: string }>) => (
                    <CampaignArea {...context} campaignID={routeComponentProps.match.params.campaignID} />
                )}
            />
        </Switch>
    )
}
