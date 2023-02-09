import React, { FC } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Routes, Route, useLocation } from 'react-router-dom-v5-compat'

import { useQuery } from '@sourcegraph/http-client'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import {
    OrgAreaOrganizationFields,
    OrgFeatureFlagValueResult,
    OrgFeatureFlagValueVariables,
} from '../../graphql-operations'
import { RouteV6Descriptor } from '../../util/contributions'
import { OrgAreaRouteContext } from '../area/OrgArea'
import { GET_ORG_FEATURE_FLAG_VALUE, ORG_DELETION_FEATURE_FLAG_NAME } from '../backend'

import { OrgSettingsSidebar, OrgSettingsSidebarItems } from './OrgSettingsSidebar'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

export interface OrgSettingsAreaRoute extends RouteV6Descriptor<OrgSettingsAreaRouteContext> {}

export interface OrgSettingsAreaProps extends OrgAreaRouteContext, ThemeProps {
    authenticatedUser: AuthenticatedUser
    sideBarItems: OrgSettingsSidebarItems
    routes: readonly OrgSettingsAreaRoute[]
    org: OrgAreaOrganizationFields
}

export interface OrgSettingsAreaRouteContext extends OrgSettingsAreaProps {
    org: OrgAreaOrganizationFields
    showOrgDeletion: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export const AuthenticatedOrgSettingsArea: FC<OrgSettingsAreaProps> = props => {
    const location = useLocation()
    const orgDeletionFlag = useQuery<OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables>(
        GET_ORG_FEATURE_FLAG_VALUE,
        {
            variables: { orgID: props.org.id, flagName: ORG_DELETION_FEATURE_FLAG_NAME },
            fetchPolicy: 'cache-and-network',
            skip: !props.authenticatedUser || !props.org.id,
        }
    )

    const context: OrgSettingsAreaRouteContext = {
        ...props,
        showOrgDeletion: orgDeletionFlag.data?.organizationFeatureFlagValue || false,
    }

    return (
        <div className="d-flex flex-column flex-sm-row">
            <OrgSettingsSidebar items={props.sideBarItems} {...context} className="flex-0 mr-3 mb-4" />
            <div className="flex-1">
                <ErrorBoundary location={location}>
                    <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
                        <Routes>
                            {props.routes.map(
                                ({ path, render, condition = () => true }) =>
                                    condition(context) && (
                                        <Route
                                            element={render(context)}
                                            path={path}
                                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        />
                                    )
                            )}
                            <Route element={<NotFoundPage />} />
                        </Routes>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        </div>
    )
}

export const OrgSettingsArea = withAuthenticatedUser(AuthenticatedOrgSettingsArea)
