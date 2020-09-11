import React from 'react'
import { RouteComponentProps, Switch, Route } from 'react-router'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignsDotComPage } from './marketing/CampaignsDotComPage'
import { AuthenticatedUser } from '../../../auth'
import { CampaignListPage } from '../list/CampaignListPage'
import { CreateCampaignPage } from '../create/CreateCampaignPage'

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

export const AuthenticatedCampaignsArea = withAuthenticatedUser<AuthenticatedProps>(({ match, ...outerProps }) => (
    <>
        {/* eslint-disable react/jsx-no-bind */}
        <Switch>
            <Route render={props => <CampaignListPage {...outerProps} {...props} />} path={match.url} exact={true} />
            <Route
                path={`${match.url}/create`}
                render={props => <CreateCampaignPage {...outerProps} {...props} />}
                exact={true}
            />
        </Switch>
        {/* eslint-enable react/jsx-no-bind */}
    </>
))
