import * as React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { BatchChangesProps } from '../batches'
import { BreadcrumbsProps, BreadcrumbSetters } from '../components/Breadcrumbs'
import { HeroPage } from '../components/HeroPage'
import { FeatureFlagProps } from '../featureFlags/featureFlags'

import { OrgArea, OrgAreaRoute } from './area/OrgArea'
import { OrgAreaHeaderNavItem } from './area/OrgHeader'
import { OrgInvitationPage } from './invitations/OrgInvitationPage'
import { NewOrganizationPage } from './new/NewOrganizationPage'
import { JoinOpenBetaPage } from './openBeta/JoinOpenBetaPage'
import { NewOrgOpenBetaPage } from './openBeta/NewOrganizationPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

export interface Props
    extends RouteComponentProps<{}>,
        ExtensionsControllerProps,
        PlatformContextProps,
        SettingsCascadeProps,
        FeatureFlagProps,
        ThemeProps,
        TelemetryProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        BatchChangesProps {
    orgAreaRoutes: readonly OrgAreaRoute[]
    orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[]

    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display organization-related pages.
 */
const AuthenticatedOrgsArea: React.FunctionComponent<Props> = props => (
    <Switch>
        {(!props.isSourcegraphDotCom || props.authenticatedUser.siteAdmin) && (
            <Route path={`${props.match.url}/new`} component={NewOrganizationPage} exact={true} />
        )}
        {props.featureFlags.get('open-beta-enabled') && (
            <Route
                path={`${props.match.url}/joinopenbeta`}
                exact={true}
                render={routeComponentProps => <JoinOpenBetaPage {...props} {...routeComponentProps} />}
            />
        )}
        {props.featureFlags.get('open-beta-enabled') && (
            <Route
                path={`${props.match.url}/joinopenbeta/neworg/:openBetaId`}
                exact={true}
                render={routeComponentProps => <NewOrgOpenBetaPage {...props} {...routeComponentProps} />}
            />
        )}
        <Route
            path={`${props.match.url}/invitation/:token`}
            exact={true}
            render={routeComponentProps => <OrgInvitationPage {...props} {...routeComponentProps} />}
        />
        <Route
            path={`${props.match.url}/:name`}
            render={routeComponentProps => <OrgArea {...props} {...routeComponentProps} />}
        />

        <Route component={NotFoundPage} />
    </Switch>
)

export const OrgsArea = withAuthenticatedUser(AuthenticatedOrgsArea)
