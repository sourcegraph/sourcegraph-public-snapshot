import * as React from 'react'

import { Route, Routes } from 'react-router-dom'

import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import type { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import type { BatchChangesProps } from '../batches'
import type { BreadcrumbSetters, BreadcrumbsProps } from '../components/Breadcrumbs'
import { NotFoundPage } from '../components/HeroPage'

import { OrgArea, type OrgAreaRoute } from './area/OrgArea'
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
}

/**
 * Renders a layout of a sidebar and a content area to display organization-related pages.
 */
const AuthenticatedOrgsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => (
    <Routes>
        <Route
            path="new"
            element={<NewOrganizationPage telemetryRecorder={props.platformContext.telemetryRecorder} />}
        />
        <Route
            path="invitation/:token"
            element={<OrgInvitationPage {...props} telemetryRecorder={props.platformContext.telemetryRecorder} />}
        />
        <Route path=":orgName/*" element={<OrgArea {...props} />} />
        <Route path="*" element={<NotFoundPage pageType="organization" />} />
    </Routes>
)

export const OrgsArea = withAuthenticatedUser(AuthenticatedOrgsArea)
