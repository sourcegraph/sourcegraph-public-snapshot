import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { SettingsArea } from '../../settings/SettingsArea'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { SettingsRepositoriesPage } from '../../user/settings/repositories/SettingsRepositoriesPage'
import { UserSettingsManageRepositoriesPage } from '../../user/settings/repositories/UserSettingsManageRepositoriesPage'
import { OrgAreaPageProps } from '../area/OrgArea'

import { OrgAddCodeHostsPageContainer } from './codeHosts/OrgAddCodeHostsPageContainer'
import { OrgSettingsMembersPage } from './members/OrgSettingsMembersPage'
import { OrgSettingsSidebar } from './OrgSettingsSidebar'
import { OrgSettingsProfilePage } from './profile/OrgSettingsProfilePage'
import { showOrganizationsCode } from './utils'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

interface Props extends OrgAreaPageProps, RouteComponentProps<{}>, ThemeProps {
    location: H.Location
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export const OrgSettingsArea: React.FunctionComponent<Props> = props => {
    if (!props.authenticatedUser) {
        return null
    }

    const showOrgsCode = showOrganizationsCode(props.authenticatedUser)

    return (
        <div className="d-flex">
            <OrgSettingsSidebar {...props} className="flex-0 mr-3" />
            <div className="flex-1">
                <ErrorBoundary location={props.location}>
                    <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                        <Switch>
                            <Route
                                path={props.match.path}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                render={routeComponentProps => (
                                    <SettingsArea
                                        {...routeComponentProps}
                                        {...props}
                                        subject={props.org}
                                        extraHeader={
                                            <>
                                                {props.authenticatedUser &&
                                                    props.org.viewerCanAdminister &&
                                                    !props.org.viewerIsMember && (
                                                        <SiteAdminAlert className="sidebar__alert">
                                                            Viewing settings for <strong>{props.org.name}</strong>
                                                        </SiteAdminAlert>
                                                    )}
                                                <p>
                                                    Organization settings apply to all members. User settings override
                                                    organization settings.
                                                </p>
                                            </>
                                        }
                                    />
                                )}
                            />
                            <Route
                                path={`${props.match.path}/profile`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                render={routeComponentProps => (
                                    <OrgSettingsProfilePage {...routeComponentProps} {...props} />
                                )}
                            />
                            <Route
                                path={`${props.match.path}/members`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                render={routeComponentProps => (
                                    <OrgSettingsMembersPage {...routeComponentProps} {...props} />
                                )}
                            />
                            {showOrgsCode && (
                                <>
                                    <Route
                                        path={`${props.match.path}/code-hosts`}
                                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        exact={true}
                                        render={routeComponentProps => (
                                            <OrgAddCodeHostsPageContainer
                                                {...routeComponentProps}
                                                owner={{
                                                    id: props.org.id,
                                                    type: 'org',
                                                    name: props.org.displayName || props.org.name,
                                                }}
                                                context={window.context}
                                                routingPrefix={`${props.org.url}/settings`}
                                                telemetryService={props.telemetryService}
                                                onUserExternalServicesOrRepositoriesUpdate={() => {}}
                                            />
                                        )}
                                    />
                                    <Route
                                        path={`${props.match.path}/repositories`}
                                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        exact={true}
                                        render={routeComponentProps => (
                                            <SettingsRepositoriesPage
                                                {...routeComponentProps}
                                                {...props}
                                                owner={{
                                                    id: props.org.id,
                                                    type: 'org',
                                                    name: props.org.displayName || props.org.name,
                                                }}
                                                routingPrefix={`${props.org.url}/settings`}
                                                onUserExternalServicesOrRepositoriesUpdate={() => {}} // TODO...
                                            />
                                        )}
                                    />
                                    <Route
                                        path={`${props.match.path}/repositories/manage`}
                                        key="hardcoded-key"
                                        exact={true}
                                        render={routeComponentProps => (
                                            <UserSettingsManageRepositoriesPage
                                                {...routeComponentProps}
                                                {...props}
                                                owner={{
                                                    id: props.org.id,
                                                    type: 'org',
                                                    name: props.org.displayName || props.org.name,
                                                }}
                                                routingPrefix={`${props.org.url}/settings`}
                                                onSyncedPublicRepositoriesUpdate={() => {}}
                                            />
                                        )}
                                    />
                                </>
                            )}
                            <Route component={NotFoundPage} />
                        </Switch>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        </div>
    )
}
