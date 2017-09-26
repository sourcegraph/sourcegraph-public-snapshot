import * as React from 'react'
import { Redirect } from 'react-router-dom'
import { ParsedRouteProps } from '../util/routes'
import { SettingsNotFoundPage } from './'
import { EditorAuthPage } from './auth/EditorAuthPage'
import { SettingsSidebar } from './SettingsSidebar'
import { AcceptInvitePage } from './team/AcceptInvitePage'
import { NewTeam } from './team/NewTeam'
import { Team } from './team/Team'
// import { UserProfilePage } from './user/UserProfilePage'

/**
 * Renders a layout of a sidebar and a content area to display different settings
 */
export class SettingsPage extends React.Component<ParsedRouteProps> {
    public render(): JSX.Element | null {
        // If not logged in, redirect to sign in
        if (!window.context.user) {
            return <Redirect to='/sign-in' />
        }
        let content: JSX.Element | null = null
        // TODO This routing should not use `this.props.routeName`, but use a React Router <Switch> element to match the subroutes
        switch (this.props.routeName) {
            case 'user-profile':
                // Don't show any ProfilePage for now
                // content = <UserProfilePage />
                content = null
                break
            case 'accept-invite':
                content = <AcceptInvitePage {...this.props}/>
                break
            case 'editor-auth':
                content = <EditorAuthPage />
                break
            case 'settings-error':
                content = <SettingsNotFoundPage />
                break
            case 'team-profile': {
                // TODO use React router to get the param
                const teamName = this.props.location.pathname.match(/^\/settings\/team\/([a-zA-Z0-9][a-zA-Z0-9\-]*)\/?$/)![1]
                content = <Team {...this.props} teamName={teamName} />
                break
            }
            case 'teams-new':
                content = <NewTeam {...this.props} />
                break
            default:
                throw new Error('Non-settings routes not supported')
        }

        return (
            <div className='settings-page'>
                <SettingsSidebar {...this.props} />
                <div className='settings-page__content'>
                    {content}
                </div>
            </div>
        )
    }
}
