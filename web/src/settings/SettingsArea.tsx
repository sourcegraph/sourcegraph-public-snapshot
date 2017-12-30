import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as H from 'history'
import * as React from 'react'
import { match, Route, Switch } from 'react-router'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { EditorAuthPage } from './EditorAuthPage'
import { AcceptInvitePage } from './org/AcceptInvitePage'
import { NewOrg } from './org/NewOrg'
import { OrgSettingsProfilePage } from './org/OrgSettingsProfilePage'
import { SettingsSidebar } from './SettingsSidebar'
import { UserSettingsProfilePage } from './user/UserSettingsProfilePage'

const SettingsNotFoundPage = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested settings page was not found."
    />
)

interface Props {
    history: H.History
    location: H.Location
    match: match<{}>
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

        return (
            <div className="settings-area">
                <SettingsSidebar history={this.props.history} location={this.props.location} user={this.props.user} />
                <div className="settings-area__content">
                    <Switch>
                        {/* Render empty page if no settings page selected */}
                        <Route
                            path={`${this.props.match.url}/profile`}
                            component={UserSettingsProfilePage}
                            exact={true}
                        />
                        <Route
                            path={`${this.props.match.url}/accept-invite`}
                            component={AcceptInvitePage}
                            exact={true}
                        />
                        <Route path={`${this.props.match.url}/editor-auth`} component={EditorAuthPage} exact={true} />
                        <Route path={`${this.props.match.url}/orgs/new`} component={NewOrg} exact={true} />
                        <Route
                            path={`${this.props.match.url}/orgs/:orgName`}
                            component={OrgSettingsProfilePage}
                            exact={true}
                        />
                        <Route component={SettingsNotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }
}
