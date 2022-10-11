import * as React from 'react'

import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables } from '../../graphql-operations'
import { SettingsArea } from '../../settings/SettingsArea'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { OrgAreaPageProps } from '../area/OrgArea'
import { GET_ORG_FEATURE_FLAG_VALUE, ORG_DELETION_FEATURE_FLAG_NAME } from '../backend'

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

interface Props extends OrgAreaPageProps, RouteComponentProps<{}>, ThemeProps {
    location: H.Location
    authenticatedUser: AuthenticatedUser
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export const OrgSettingsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const orgDeletionFlag = useQuery<OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables>(
        GET_ORG_FEATURE_FLAG_VALUE,
        {
            variables: { orgID: props.org.id, flagName: ORG_DELETION_FEATURE_FLAG_NAME },
            fetchPolicy: 'cache-and-network',
            skip: !props.authenticatedUser || !props.org.id,
        }
    )

    if (!props.authenticatedUser) {
        return null
    }

    const showOrgDeletion = orgDeletionFlag.data?.organizationFeatureFlagValue || false

    return (
        <div className="d-flex">
            <OrgSettingsSidebar {...props} className="flex-0 mr-3" />
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
                                                    <Text>
                                                        Organization settings apply to all members. User settings
                                                        override organization settings.
                                                    </Text>
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
                            <Route component={NotFoundPage} />
                        </Switch>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        </div>
    )
}
