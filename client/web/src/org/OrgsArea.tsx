import * as React from 'react'

import { Routes, Route, useParams, useLocation, useNavigate } from 'react-router-dom'

import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import type { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import type { BatchChangesProps } from '../batches'
import type { BreadcrumbsProps, BreadcrumbSetters } from '../components/Breadcrumbs'
import { NotFoundPage } from '../components/HeroPage'

import { OrgArea, type OrgAreaProps, type OrgAreaRoute } from './area/OrgArea'
import type { OrgAreaHeaderNavItem } from './area/OrgHeader'
import { OrgInvitationPage } from './invitations/OrgInvitationPage'
import { NewOrganizationPage } from './new/NewOrganizationPage'
import type { OrgSettingsAreaRoute } from './settings/OrgSettingsArea'
import type { OrgSettingsSidebarItems } from './settings/OrgSettingsSidebar'

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
        <Route path="*" element={<NotFoundPage pageType="organization" />} />
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
