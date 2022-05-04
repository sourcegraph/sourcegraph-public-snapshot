import * as React from 'react'

import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables } from '../../graphql-operations'
import { SettingsArea } from '../../settings/SettingsArea'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { SettingsRepositoriesPage } from '../../user/settings/repositories/SettingsRepositoriesPage'
import { UserSettingsManageRepositoriesPage } from '../../user/settings/repositories/UserSettingsManageRepositoriesPage'
import { OrgAreaPageProps } from '../area/OrgArea'
import { ORG_CODE_FEATURE_FLAG_NAME, GET_ORG_FEATURE_FLAG_VALUE, ORG_DELETION_FEATURE_FLAG_NAME } from '../backend'
import { useEventBus } from '../emitter'

import { OrgAddCodeHostsPageContainer } from './codeHosts/OrgAddCodeHostsPageContainer'
import { DeleteOrg } from './DeleteOrg'
import { OrgSettingsMembersPage } from './members-v1/OrgSettingsMembersPage'
import { OrgSettingsSidebar } from './OrgSettingsSidebar'
import { OrgSettingsProfilePage } from './profile/OrgSettingsProfilePage'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

const LoadingComponent: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <LoadingSpinner className="m-2" />
)

interface Props extends OrgAreaPageProps, RouteComponentProps<{}>, ThemeProps {
    location: H.Location
    authenticatedUser: AuthenticatedUser
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export const OrgSettingsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const emitter = useEventBus()
    // we can ignore the error states in this case
    // if there is an error, we will not show the code host connections and repository screens
    // same for until the feature flag value is loaded (which in practice should be fast)
    const { data, loading } = useQuery<OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables>(
        GET_ORG_FEATURE_FLAG_VALUE,
        {
            variables: { orgID: props.org.id, flagName: ORG_CODE_FEATURE_FLAG_NAME },
            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',
            skip: !props.authenticatedUser || !props.org.id,
        }
    )

    const orgDeletionFlag = useQuery<OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables>(
        GET_ORG_FEATURE_FLAG_VALUE,
        {
            variables: { orgID: props.org.id, flagName: ORG_DELETION_FEATURE_FLAG_NAME },
            fetchPolicy: 'cache-and-network',
            skip: !props.authenticatedUser || !props.org.id,
        }
    )

    const onOrgGetStartedRefresh = (): void => {
        emitter.emit('refreshOrgHeader', 'refreshing due to changes on repo setup')
    }

    if (!props.authenticatedUser) {
        return null
    }

    const showOrgCode = data?.organizationFeatureFlagValue || false
    const showOrgDeletion = orgDeletionFlag.data?.organizationFeatureFlagValue || false

    return (
        <div className="d-flex">
            <OrgSettingsSidebar {...props} className="flex-0 mr-3" showOrgCode={showOrgCode} />
            <div className="flex-1">
                <ErrorBoundary location={props.location}>
                    <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
                        <Switch>
                            <Route
                                path={props.match.path}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                render={routeComponentProps => (
                                    <div>
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
                                                        Organization settings apply to all members. User settings
                                                        override organization settings.
                                                    </p>
                                                </>
                                            }
                                        />
                                        {props.isSourcegraphDotCom && props.org.viewerIsMember && showOrgDeletion && (
                                            <DeleteOrg {...routeComponentProps} {...props} />
                                        )}
                                    </div>
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
                            {!props.newMembersInviteEnabled && (
                                <Route
                                    path={`${props.match.path}/members`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={true}
                                    render={routeComponentProps => (
                                        <OrgSettingsMembersPage {...routeComponentProps} {...props} />
                                    )}
                                />
                            )}
                            {showOrgCode && [
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
                                            onOrgGetStartedRefresh={onOrgGetStartedRefresh}
                                            context={window.context}
                                            routingPrefix={`${props.org.url}/settings`}
                                            telemetryService={props.telemetryService}
                                            onUserExternalServicesOrRepositoriesUpdate={() => {}}
                                        />
                                    )}
                                />,
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
                                            onOrgGetStartedRefresh={onOrgGetStartedRefresh}
                                            routingPrefix={`${props.org.url}/settings`}
                                            onUserExternalServicesOrRepositoriesUpdate={() => {}} // TODO...
                                        />
                                    )}
                                />,
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
                                />,
                            ]}
                            <Route component={loading ? LoadingComponent : NotFoundPage} />
                        </Switch>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        </div>
    )
}
