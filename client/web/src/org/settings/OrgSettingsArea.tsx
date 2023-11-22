import React, { type FC } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Routes, Route } from 'react-router-dom'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { RouteError } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import type { OrgAreaOrganizationFields } from '../../graphql-operations'
import type { RouteV6Descriptor } from '../../util/contributions'
import type { OrgAreaRouteContext } from '../area/OrgArea'

import { OrgSettingsSidebar, type OrgSettingsSidebarItems } from './OrgSettingsSidebar'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

export interface OrgSettingsAreaRoute extends RouteV6Descriptor<OrgSettingsAreaRouteContext> {}

export interface OrgSettingsAreaProps extends OrgAreaRouteContext {
    authenticatedUser: AuthenticatedUser
    sideBarItems: OrgSettingsSidebarItems
    routes: readonly OrgSettingsAreaRoute[]
    org: OrgAreaOrganizationFields
}

export interface OrgSettingsAreaRouteContext extends OrgSettingsAreaProps {
    org: OrgAreaOrganizationFields
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export const AuthenticatedOrgSettingsArea: FC<OrgSettingsAreaProps> = props => {
    const context: OrgSettingsAreaRouteContext = {
        ...props,
    }

    return (
        <div className="d-flex flex-column flex-sm-row">
            <OrgSettingsSidebar items={props.sideBarItems} {...context} className="flex-0 mr-3 mb-4" />
            <div className="flex-1">
                <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
                    <Routes>
                        {props.routes.map(
                            ({ path, render, condition = () => true }) =>
                                condition(context) && (
                                    <Route
                                        element={render(context)}
                                        errorElement={<RouteError />}
                                        path={path}
                                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    />
                                )
                        )}
                        <Route path="*" element={<NotFoundPage />} />
                    </Routes>
                </React.Suspense>
            </div>
        </div>
    )
}

export const OrgSettingsArea = withAuthenticatedUser(AuthenticatedOrgSettingsArea)
