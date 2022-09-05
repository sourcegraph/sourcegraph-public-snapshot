import * as React from 'react'

import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { OrgAreaPageProps } from '../area/OrgArea'
import { useEventBus } from '../emitter'

import { OrgMembersListPage } from './OrgMembersListPage'
import { OrgMembersSidebar } from './OrgMembersSidebar'
import { OrgPendingInvitesPage } from './OrgPendingInvites'

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
export const OrgMembersArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const emitter = useEventBus()
    if (!props.authenticatedUser) {
        return null
    }

    const onOrgGetStartedRefresh = (): void => {
        emitter.emit('refreshOrgHeader', 'refreshing due to changes on members section')
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
                                    <OrgMembersListPage
                                        key={props.org.name}
                                        {...routeComponentProps}
                                        {...props}
                                        onOrgGetStartedRefresh={onOrgGetStartedRefresh}
                                    />
                                )}
                            />
                            <Route
                                path={`${props.match.path}/pending-invites`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={true}
                                render={routeComponentProps => (
                                    <OrgPendingInvitesPage
                                        key={props.org.name}
                                        {...routeComponentProps}
                                        {...props}
                                        onOrgGetStartedRefresh={onOrgGetStartedRefresh}
                                    />
                                )}
                            />
                            <Route component={NotFoundPage} />
                        </Switch>
                    </React.Suspense>
                </ErrorBoundary>
            </div>
        </div>
    )
}
