import * as React from 'react'

import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { OrgAreaOrganizationFields } from '../../graphql-operations'
import { RouteDescriptor } from '../../util/contributions'
import { OrgAreaRouteContext } from '../area/OrgArea'

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
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export const AuthenticatedOrgSettingsArea: React.FunctionComponent<
    React.PropsWithChildren<OrgSettingsAreaProps>
> = props => {
    const context: OrgSettingsAreaRouteContext = {
        ...props,
    }

    return (
        <div className="d-flex">
            <OrgSettingsSidebar items={props.sideBarItems} {...context} className="flex-0 mr-3" />
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
