import * as React from 'react'
import { Redirect } from 'react-router-dom'
import { viewEvents } from '../tracking/events'
import { sourcegraphContext } from '../util/sourcegraphContext'
import { EditorAuthPage } from './auth/EditorAuthPage'
import { SignInPage } from './auth/SignInPage'
import { UserProfilePage } from './user/UserProfilePage'

export type SettingsSelection = 'editor-auth' | 'sign-in' | 'user-profile'

export interface Props {
    routeName: SettingsSelection
}

export class SettingsPage extends React.Component<Props> {
    public render(): JSX.Element | null {
        let content: JSX.Element | null = null
        switch (this.props.routeName) {
            case 'user-profile':
                viewEvents.UserProfile.log()
                content = sourcegraphContext.user ? <UserProfilePage /> : <SignInPage showEditorFlow={false} />
                break
            case 'editor-auth':
                viewEvents.EditorAuth.log()
                content = sourcegraphContext.user ? <EditorAuthPage /> : <SignInPage showEditorFlow={true} />
                break
            case 'sign-in':
                viewEvents.SignIn.log()
                content = sourcegraphContext.user ? <Redirect to='/search' /> : <SignInPage showEditorFlow={false} />
                break
        }

        return (
            <div className='settings-page'>
                {/* <SettingsSidebar {...this.props} /> */}
                <div className='settings-page__content'>
                    {content}
                </div>
            </div>
        )
    }
}
