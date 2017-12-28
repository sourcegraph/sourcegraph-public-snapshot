import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as H from 'history'
import * as React from 'react'
import { match, Route, RouteProps, Switch } from 'react-router'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { SiteAdminAllUsersPage } from './SiteAdminAllUsersPage'
import { SiteAdminAnalyticsPage } from './SiteAdminAnalyticsPage'
import { SiteAdminConfigurationPage } from './SiteAdminConfigurationPage'
import { SiteAdminOrgsPage } from './SiteAdminOrgsPage'
import { SiteAdminOverviewPage } from './SiteAdminOverviewPage'
import { SiteAdminRepositoriesPage } from './SiteAdminRepositoriesPage'
import { SiteAdminSidebar } from './SiteAdminSidebar'
import { SiteAdminTelemetryPage } from './SiteAdminTelemetryPage'

const NotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested site admin page was not found."
    />
)

interface SettingsPageProps {
    history: H.History
    location: H.Location
    match: match<{}>
    user: GQL.IUser | null
}

/**
 * Renders a layout of a sidebar and a content area to display site admin information.
 */
export class SiteAdminArea extends React.Component<SettingsPageProps> {
    public render(): JSX.Element | null {
        // If not logged in, redirect to sign in.
        if (!this.props.user) {
            const newUrl = new URL(window.location.href)
            newUrl.pathname = '/sign-in'
            // Return to the current page after sign up/in.
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={newUrl.pathname + newUrl.search} />
        }

        // Transfer the user prop to the routes' components.
        const RouteWithProps = (props: RouteProps): React.ReactElement<Route> => (
            <Route
                {...props}
                component={undefined}
                // tslint:disable-next-line:jsx-no-lambda
                render={props2 => {
                    const finalProps = { ...props2, user: this.props.user }
                    if (props.component) {
                        return React.createElement(props.component, finalProps)
                    }
                    if (props.render) {
                        return props.render(finalProps)
                    }
                    return null
                }}
            />
        )

        return (
            <div className="site-admin-area">
                <SiteAdminSidebar history={this.props.history} location={this.props.location} />
                <div className="site-admin-area__content">
                    <Switch>
                        {/* Render empty page if no page selected. */}
                        <RouteWithProps path={this.props.match.url} component={SiteAdminOverviewPage} exact={true} />
                        <RouteWithProps
                            path={`${this.props.match.url}/configuration`}
                            component={SiteAdminConfigurationPage}
                            exact={true}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/repositories`}
                            component={SiteAdminRepositoriesPage}
                            exact={true}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/organizations`}
                            component={SiteAdminOrgsPage}
                            exact={true}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/users`}
                            component={SiteAdminAllUsersPage}
                            exact={true}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/analytics`}
                            component={SiteAdminAnalyticsPage}
                            exact={true}
                        />
                        <RouteWithProps
                            path={`${this.props.match.url}/telemetry`}
                            component={SiteAdminTelemetryPage}
                            exact={true}
                        />
                        <RouteWithProps component={NotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }
}
