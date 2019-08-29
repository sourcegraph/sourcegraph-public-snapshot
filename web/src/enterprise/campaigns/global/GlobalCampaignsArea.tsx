import React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { GlobalCampaignListPage } from './list/GlobalCampaignListPage'

interface Props
    extends RouteComponentProps<{}>,
        Pick<
            NamespaceCampaignsAreaContext,
            Exclude<keyof NamespaceCampaignsAreaContext, 'campaignsURL' | 'namespace' | 'setBreadcrumbItem'>
        > {}

/**
 * The global campaigns area.
 */
export const GlobalCampaignsArea: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const context: Pick<NamespaceCampaignsAreaContext, Exclude<keyof NamespaceCampaignsAreaContext, 'namespace'>> = {
        ...props,
        campaignsURL: match.url,
    }
    return (
        <Switch>
            <Route path={context.campaignsURL} exact={true}>
                <div className="container mt-4">
                    <GlobalCampaignListPage {...context} />
                </div>
            </Route>
        </Switch>
    )
}
