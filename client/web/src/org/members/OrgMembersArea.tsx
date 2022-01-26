import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { OrgFeatureFlagValueResult, OrgFeatureFlagValueVariables } from '../../graphql-operations'
import { OrgAreaPageProps } from '../area/OrgArea'
import { ORG_CODE_FEATURE_FLAG_NAME, GET_ORG_FEATURE_FLAG_VALUE } from '../backend'
import { OrgMembersSidebar } from './OrgMembersSidebar'
import { OrgMembersListPage } from './OrgMembersListPage'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

const LoadingComponent: React.FunctionComponent = () => <LoadingSpinner className="m-2" />

interface Props extends OrgAreaPageProps, RouteComponentProps<{}>, ThemeProps {
    location: H.Location
    authenticatedUser: AuthenticatedUser
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export const OrgMembersArea: React.FunctionComponent<Props> = props => {
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

    if (!props.authenticatedUser) {
        return null
    }
    return (
        <div className="d-flex">
            <OrgMembersSidebar {...props} className="flex-0 mr-3" />
            <div className="flex-1">
                <ErrorBoundary location={props.location}>
                    <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
                        <Switch>
                            <Route
                                path={props.match.path}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                render={routeComponentProps => (
                                    <OrgMembersListPage {...routeComponentProps} {...props} />
                                )}
                            />
                            <Route
                                path={`${props.match.path}/pending-invites`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                render={routeComponentProps => <div>pending invites</div>}
                            />
                            <Route component={loading ? LoadingComponent : NotFoundPage} />
                        </Switch>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        </div>
    )
}
