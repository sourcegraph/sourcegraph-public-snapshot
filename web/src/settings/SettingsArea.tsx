import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { AcceptInvitePage } from '../org/AcceptInvitePage'
import { EditorAuthPage } from '../user/settings/EditorAuthPage'
import { UserSettingsAccountPage } from '../user/settings/UserSettingsAccountPage'
import { UserSettingsConfigurationPage } from '../user/settings/UserSettingsConfigurationPage'
import { UserSettingsEmailsPage } from '../user/settings/UserSettingsEmailsPage'
import { UserSettingsProfilePage } from '../user/settings/UserSettingsProfilePage'
import { SettingsSidebar } from './SettingsSidebar'

const SettingsNotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested settings page was not found."
    />
)

interface Props extends RouteComponentProps<{}> {
    user: GQL.IUser | null
}

/**
 * Renders a layout of a sidebar and a content area to display settings.
 */
export class SettingsArea extends React.Component<Props> {
    public render(): JSX.Element | null {
        // If not logged in, redirect to sign in
        if (!this.props.user) {
            const currUrl = new URL(window.location.href)
            const newUrl = new URL(window.location.href)
            newUrl.pathname = currUrl.pathname === '/settings/accept-invite' ? '/sign-up' : '/sign-in'
            // Return to the current page after sign up/in.
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={newUrl.pathname + newUrl.search} />
        }

        if (this.props.location.pathname === '/settings') {
            return <Redirect to="/settings/profile" />
        }

        const transferProps = { user: this.props.user }

        return (
            <div className="settings-area area">
                <SettingsSidebar history={this.props.history} location={this.props.location} user={this.props.user} />
                <div className="area__content">
                    <Switch>
                        {/* Render empty page if no settings page selected */}
                        <Route
                            path={`${this.props.match.url}/profile`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <UserSettingsProfilePage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/configuration`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <UserSettingsConfigurationPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/account`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <UserSettingsAccountPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/emails`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <UserSettingsEmailsPage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/accept-invite`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <AcceptInvitePage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${this.props.match.url}/editor-auth`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            component={EditorAuthPage}
                        />
                        <Route component={SettingsNotFoundPage} key="hardcoded-key" />
                    </Switch>
                </div>
            </div>
        )
    }
}
