import React from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'
import { GlobalCampaignListPage } from './list/GlobalCampaignListPage'
import { IUser } from '../../../../../shared/src/graphql/schema'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignsDotComPage } from './marketing/CampaignsDotComPage'
import { CampaignsSiteAdminMarketingPage } from './marketing/CampaignsSiteAdminMarketingPage'
import { CampaignsUserMarketingPage } from './marketing/CampaignsUserMarketingPage'
import { DismissibleAlert } from '../../../components/DismissibleAlert'
import { CampaignArea } from '../detail/CampaignArea'
import { NewCampaignPage } from '../new/NewCampaignPage'

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
    if (outerProps.isSourcegraphDotCom) {
        return (
            <div className="container mt-4">
                <CampaignsDotComPage {...outerProps} />
            </div>
        )
    }
    if (window.context.experimentalFeatures?.automation === 'enabled') {
        if (!outerProps.authenticatedUser.siteAdmin && window.context.site['campaigns.readAccess.enabled'] !== true) {
            return (
                <div className="container mt-4">
                    <CampaignsUserMarketingPage {...outerProps} enableReadAccess={true} />
                </div>
            )
        }
        return (
            <>
                <DismissibleAlert partialStorageKey="campaigns-beta" className="alert-info">
                    <p className="mb-0">
                        Campaigns are currently in beta. During the beta period, campaigns are free to use. After the
                        beta period, campaigns will be available as a paid add-on. Get in touch on Twitter{' '}
                        <a href="https://twitter.com/srcgraph">@srcgraph</a>, file an issue in our{' '}
                        <a href="https://github.com/sourcegraph/sourcegraph/issues">public issue tracker</a>, or email{' '}
                        <a href="mailto:feedback@sourcegraph.com?subject=Feedback on Campaigns">
                            feedback@sourcegraph.com
                        </a>
                        . We're looking forward to your feedback!
                    </p>
                </DismissibleAlert>
                {/* eslint-disable react/jsx-no-bind */}
                <Switch>
                    <Route
                        render={props => (
                            <div className="container mt-4">
                                <GlobalCampaignListPage {...outerProps} {...props} />
                            </div>
                        )}
                        path={match.url}
                        exact={true}
                    />
                    <Route
                        path={`${match.url}/new`}
                        render={props => (
                            <div className="container mt-4">
                                <NewCampaignPage {...outerProps} {...props} />
                            </div>
                        )}
                        exact={true}
                    />
                    <Route
                        path={`${match.url}/:campaignID`}
                        render={({ match, ...props }: RouteComponentProps<{ campaignID: string }>) => (
                            <CampaignArea
                                {...outerProps}
                                {...props}
                                match={match}
                                campaignID={match.params.campaignID}
                            />
                        )}
                    />
                </Switch>
                {/* eslint-enable react/jsx-no-bind */}
            </>
        )
    }

    if (outerProps.authenticatedUser.siteAdmin) {
        return (
            <div className="container mt-4">
                <CampaignsSiteAdminMarketingPage {...outerProps} />
            </div>
        )
    }

    return (
        <div className="container mt-4">
            <CampaignsUserMarketingPage {...outerProps} enableReadAccess={false} />
        </div>
    )
})
