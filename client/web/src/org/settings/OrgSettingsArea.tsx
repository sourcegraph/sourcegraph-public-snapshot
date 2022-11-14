import * as React from 'react'

import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

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
import { RouteDescriptor } from '../../util/contributions'
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

export interface OrgSettingsAreaRoute extends RouteDescriptor<OrgSettingsAreaRouteContext> {}

export interface OrgSettingsAreaProps extends OrgAreaRouteContext, RouteComponentProps<{}>, ThemeProps {
    location: H.Location
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
export const AuthenticatedOrgSettingsArea: React.FunctionComponent<
    React.PropsWithChildren<OrgSettingsAreaProps>
> = props => {
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
                <ErrorBoundary location={props.location}>
                    <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
                        <Switch>
                            {props.routes.map(
                                ({ path, exact, render, condition = () => true }) =>
                                    condition(context) && (
                                        <Route
                                            render={routeComponentProps =>
                                                render({ ...context, ...routeComponentProps })
                                            }
                                            path={props.match.url + path}
                                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                            exact={exact}
                                        />
                                    )
                            )}
                            <Route component={NotFoundPage} />
                        </Switch>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        </div>
    )
}

export const OrgSettingsArea = withAuthenticatedUser(AuthenticatedOrgSettingsArea)
