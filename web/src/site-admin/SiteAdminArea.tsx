import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { Route, RouteComponentProps, RouteProps, Switch } from 'react-router'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { SiteAdminAllUsersPage } from './SiteAdminAllUsersPage'
import { SiteAdminAnalyticsPage } from './SiteAdminAnalyticsPage'
import { SiteAdminConfigurationPage } from './SiteAdminConfigurationPage'
import { SiteAdminInviteUserPage } from './SiteAdminInviteUserPage'
import { SiteAdminOrgsPage } from './SiteAdminOrgsPage'
import { SiteAdminOverviewPage } from './SiteAdminOverviewPage'
import { SiteAdminRepositoriesPage } from './SiteAdminRepositoriesPage'
import { SiteAdminSidebar } from './SiteAdminSidebar'
import { SiteAdminTelemetryPage } from './SiteAdminTelemetryPage'
import { SiteAdminThreadsPage } from './SiteAdminThreadsPage'
import { SiteAdminUpdatesPage } from './SiteAdminUpdatesPage'

const NotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested site admin page was not found."
    />
)

const NotSiteAdminPage = () => (
    <HeroPage icon={DirectionalSignIcon} title="403: Forbidden" subtitle="Only site admins are allowed here." />
)

// Transfer the user prop to the routes' components.
const RouteWithProps = (props: RouteProps & { user: GQL.IUser | null }): React.ReactElement<Route> => (
    <Route
        {...props}
        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
        component={undefined}
        // tslint:disable-next-line:jsx-no-lambda
        render={props2 => {
            const finalProps: Props = { ...props2, user: props.user }
            if (props.component) {
                const C = props.component
                return <C {...finalProps} />
            }
            if (props.render) {
                return props.render(finalProps)
            }
            return null
        }}
    />
)

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser | null
}

/**
 * Renders a layout of a sidebar and a content area to display site admin information.
 */
export class SiteAdminArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        // If not logged in, redirect to sign in.
        if (!this.props.user) {
            const newUrl = new URL(window.location.href)
            newUrl.pathname = '/sign-in'
            // Return to the current page after sign up/in.
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={newUrl.pathname + newUrl.search} />
        }

        // If not site admin, redirect to sign in.
        if (!this.props.user.siteAdmin) {
            return <NotSiteAdminPage />
        }

        const transferProps = { user: this.props.user }

        return (
            <div className="site-admin-area area">
                <SiteAdminSidebar history={this.props.history} location={this.props.location} />
                <div className="area__content">
                    <Switch>
                        {/* Render empty page if no page selected. */}
                        <RouteWithProps
                            path={this.props.match.url}
                            component={SiteAdminOverviewPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/configuration`}
                            component={SiteAdminConfigurationPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/repositories`}
                            component={SiteAdminRepositoriesPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/organizations`}
                            component={SiteAdminOrgsPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/users`}
                            component={SiteAdminAllUsersPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/invite-user`}
                            component={SiteAdminInviteUserPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/threads`}
                            component={SiteAdminThreadsPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/analytics`}
                            component={SiteAdminAnalyticsPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/updates`}
                            component={SiteAdminUpdatesPage}
                            exact={true}
                            {...transferProps}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/telemetry`}
                            component={SiteAdminTelemetryPage}
                            exact={true}
                            {...transferProps}
                        />
                        <Route component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }
}
