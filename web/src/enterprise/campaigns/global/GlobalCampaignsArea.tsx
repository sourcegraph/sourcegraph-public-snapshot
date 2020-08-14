import React from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'
import { CampaignDetails } from '../detail/CampaignDetails'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { ThemeProps } from '../../../../../shared/src/theme'
import { CreateCampaign } from './create/CreateCampaign'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignCliHelp } from './create/CampaignCliHelp'
import { CampaignsDotComPage } from './marketing/CampaignsDotComPage'
import { CampaignsSiteAdminMarketingPage } from './marketing/CampaignsSiteAdminMarketingPage'
import { CampaignsUserMarketingPage } from './marketing/CampaignsUserMarketingPage'
import { AuthenticatedUser } from '../../../auth'
import { CampaignApplyPage } from '../apply/CampaignApplyPage'
import { CampaignListPage } from '../list/CampaignListPage'

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

/**
 * The global campaigns area.
 */
export const GlobalCampaignsArea: React.FunctionComponent<Props> = props => {
    if (props.isSourcegraphDotCom) {
        return (
            <div className="container web-content mt-4">
                <CampaignsDotComPage />
            </div>
        )
    }
    return (
        <div className="container web-content mt-4">
            <AuthenticatedCampaignsArea {...props} />
        </div>
    )
}

interface AuthenticatedProps extends Props {
    authenticatedUser: AuthenticatedUser
}

export const AuthenticatedCampaignsArea = withAuthenticatedUser<AuthenticatedProps>(({ match, ...outerProps }) => {
    if (window.context.experimentalFeatures?.automation === 'enabled') {
        if (!outerProps.authenticatedUser.siteAdmin && window.context.site['campaigns.readAccess.enabled'] !== true) {
            return <CampaignsUserMarketingPage enableReadAccess={true} />
        }
        return (
            <>
                {/* eslint-disable react/jsx-no-bind */}
                <Switch>
                    <Route
                        render={props => <CampaignListPage {...outerProps} {...props} />}
                        path={match.url}
                        exact={true}
                    />
                    <Route
                        path={`${match.url}/create`}
                        render={props => <CreateCampaign {...outerProps} {...props} />}
                        exact={true}
                    />
                    <Route
                        path={`${match.url}/cli`}
                        render={props => <CampaignCliHelp {...outerProps} {...props} />}
                        exact={true}
                    />
                    <Route
                        path={`${match.url}/apply/:specID`}
                        render={({ match, ...props }: RouteComponentProps<{ specID: string }>) => (
                            <CampaignApplyPage {...outerProps} {...props} specID={match.params.specID} />
                        )}
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
    if (outerProps.authenticatedUser.siteAdmin) {
        return <CampaignsSiteAdminMarketingPage />
    }
    return <CampaignsUserMarketingPage enableReadAccess={false} />
})
