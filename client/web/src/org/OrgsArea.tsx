import * as React from 'react'

import { Routes, Route, useParams, useLocation, useNavigate } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { BatchChangesProps } from '../batches'
import { BreadcrumbsProps, BreadcrumbSetters } from '../components/Breadcrumbs'
import { NotFoundPage } from '../components/HeroPage'

import { OrgArea, type OrgAreaProps, OrgAreaRoute } from './area/OrgArea'
import { OrgAreaHeaderNavItem } from './area/OrgHeader'
import { OrgInvitationPage } from './invitations/OrgInvitationPage'
import { NewOrganizationPage } from './new/NewOrganizationPage'
import { OrgSettingsAreaRoute } from './settings/OrgSettingsArea'
import { OrgSettingsSidebarItems } from './settings/OrgSettingsSidebar'

export interface Props
    extends PlatformContextProps,
        SettingsCascadeProps,
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
    isSourcegraphApp: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display organization-related pages.
 */
const AuthenticatedOrgsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => (
    <Routes>
        {(!props.isSourcegraphDotCom || props.authenticatedUser.siteAdmin) && (
            <Route path="new" element={<NewOrganizationPage />} />
        )}
        <Route path="invitation/:token" element={<OrgInvitationPage {...props} />} />
        <Route path=":orgName/*" element={<OrgAreaWithRouteProps {...props} />} />
        <Route element={<NotFoundPage pageType="organization" />} />
    </Routes>
)

// TODO: Migrate this into the OrgArea component once it's migrated to a function component.
function OrgAreaWithRouteProps(props: Omit<OrgAreaProps, 'orgName' | 'location' | 'navigate'>): JSX.Element {
    const { orgName } = useParams<{ orgName: string }>()
    const location = useLocation()
    const navigate = useNavigate()

    return <OrgArea {...props} orgName={orgName!} location={location} navigate={navigate} />
}

export const OrgsArea = withAuthenticatedUser(AuthenticatedOrgsArea)
