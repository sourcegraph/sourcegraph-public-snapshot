import React from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'
import { GlobalCampaignListPage } from './list/GlobalCampaignListPage'
import { CampaignDetails } from '../detail/CampaignDetails'
import { IUser } from '../../../../../shared/src/graphql/schema'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignUpdateSelection } from '../detail/CampaignUpdateSelection'
import { CampaignsDotComPage } from './marketing/CampaignsDotComPage'
import { CampaignsSiteAdminMarketingPage } from './marketing/CampaignsSiteAdminMarketingPage'
import { CampaignsUserMarketingPage } from './marketing/CampaignsUserMarketingPage'

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
export const GlobalCampaignsArea = withAuthenticatedUser<Props>(({ match, ...outerProps }) => {
    let content: React.ReactFragment
    if (outerProps.isSourcegraphDotCom) {
        content = <CampaignsDotComPage {...outerProps} />
    } else if (window.context.experimentalFeatures?.automation === 'enabled') {
        if (!outerProps.authenticatedUser.siteAdmin && window.context.site['campaigns.readAccess.enabled'] !== true) {
            content = <CampaignsUserMarketingPage {...outerProps} enableReadAccess={true} />
        } else {
            content = (
                <>
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
                </>
            )
        }
    } else if (outerProps.authenticatedUser.siteAdmin) {
        content = <CampaignsSiteAdminMarketingPage {...outerProps} />
    } else {
        content = <CampaignsUserMarketingPage {...outerProps} enableReadAccess={false} />
    }
    return <div className="container mt-4">{content}</div>
})
