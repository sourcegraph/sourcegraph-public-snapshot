import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as H from 'history'
import * as React from 'react'
import { match, Route, Switch } from 'react-router'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { EditorAuthPage } from './EditorAuthPage'
import { SettingsSidebar } from './SettingsSidebar'
import { AcceptInvitePage } from './team/AcceptInvitePage'
import { NewTeam } from './team/NewTeam'
import { Team } from './team/Team'
// import { UserProfilePage } from './user/UserProfilePage'

const SettingsNotFoundPage = () => <HeroPage icon={DirectionalSignIcon} title='404: Not Found' subtitle='Sorry, the requested settings page was not found.' />

interface SettingsPageProps {
    history: H.History
    match: match<{}>
}

/**
 * Renders a layout of a sidebar and a content area to display different settings
 */
export class SettingsPage extends React.Component<SettingsPageProps> {
    public render(): JSX.Element | null {
        // If not logged in, redirect to sign in
        if (!window.context.user) {
            return <Redirect to='/sign-in' />
        }
        return (
            <div className='settings-page'>
                <SettingsSidebar history={this.props.history} />
                <div className='settings-page__content'>
                    <Switch>
                        {/* Render empty page if no settings page selected */}
                        <Route path={this.props.match.url} exact={true} />
                        <Route path={`${this.props.match.url}/accept-invite`} component={AcceptInvitePage} exact={true} />
                        <Route path={`${this.props.match.url}/editor-auth`} component={EditorAuthPage} exact={true} />
                        <Route path={`${this.props.match.url}/teams/new`} component={NewTeam} exact={true} />
                        <Route path={`${this.props.match.url}/teams/:teamName`} component={Team} exact={true} />
                        <Route component={SettingsNotFoundPage} />
                    </Switch>
                </div>
            </div>
        )
    }
}
