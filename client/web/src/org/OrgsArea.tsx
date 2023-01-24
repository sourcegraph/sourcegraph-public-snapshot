import * as React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { RouteComponentProps, Switch } from 'react-router'
import { CompatRoute } from 'react-router-dom-v5-compat'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { BatchChangesProps } from '../batches'
import { BreadcrumbsProps, BreadcrumbSetters } from '../components/Breadcrumbs'
import { HeroPage } from '../components/HeroPage'

import { OrgArea, OrgAreaRoute } from './area/OrgArea'
import { OrgAreaHeaderNavItem } from './area/OrgHeader'
import { OrgInvitationPage } from './invitations/OrgInvitationPage'
import { NewOrganizationPage } from './new/NewOrganizationPage'
import { OrgSettingsAreaRoute } from './settings/OrgSettingsArea'
import { OrgSettingsSidebarItems } from './settings/OrgSettingsSidebar'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

export interface Props
    extends RouteComponentProps<{}>,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        BatchChangesProps {
    orgAreaRoutes: readonly OrgAreaRoute[]
    orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[]
    orgSettingsSideBarItems: OrgSettingsSidebarItems
    orgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[]

    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display organization-related pages.
 */
const AuthenticatedOrgsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => (
    <Switch>
        {(!props.isSourcegraphDotCom || props.authenticatedUser.siteAdmin) && (
            <CompatRoute path={`${props.match.url}/new`} component={NewOrganizationPage} exact={true} />
        )}
        <CompatRoute
            path={`${props.match.url}/invitation/:token`}
            exact={true}
            render={(routeComponentProps: RouteComponentProps<{ token: string }>) => (
                <OrgInvitationPage {...props} {...routeComponentProps} />
            )}
        />
        <CompatRoute
            path={`${props.match.url}/:name`}
            render={(routeComponentProps: RouteComponentProps<{ name: string }>) => (
                <OrgArea {...props} {...routeComponentProps} />
            )}
        />

        <CompatRoute component={NotFoundPage} />
    </Switch>
)

export const OrgsArea = withAuthenticatedUser(AuthenticatedOrgsArea)
