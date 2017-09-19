import KeyIcon from '@sourcegraph/icons/lib/Key'
import * as React from 'react'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { events } from '../../tracking/events'
import { sourcegraphContext } from '../../util/sourcegraphContext'
import { SignInButton } from './SignInButton'

interface Props {
    showEditorFlow: boolean
}

const newAuth = true

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class SignInPage extends React.Component<Props> {

    public render(): JSX.Element | null {
        let contents: JSX.Element
        if (!newAuth) { // TODO remove after Sept 20 auth switchover
            // TODO(Dan): don't just use '/' on non-editor sign ins
            // tslint:disable-next-line
            const returnTo = this.props.showEditorFlow ? '/editor-auth' : '/'
            const url = `/-/github-oauth/initiate?return-to=${returnTo}`

            contents = (
                <form method='POST' action={url} onSubmit={this.logTelemetryOnSignIn} className='settings-form'>
                    <input type='hidden' name='gorilla.csrf.Token' value={sourcegraphContext.csrfToken} />
                    <p>
                        <input type='submit' value='Sign in with GitHub' className='ui-button' />
                    </p>
                </form>
            )
        } else {
            contents = (
                <div>
                    <SignInButton />
                </div>
            )
        }

        return (
            <div className='ui-section'>
                <PageTitle title='sign in or sign up' />
                <HeroPage icon={KeyIcon} title='Welcome to Sourcegraph' subtitle='Sign in or sign up to create an account' cta={contents} />
            </div>
        )
    }

    private logTelemetryOnSignIn(): void {
        events.InitiateGitHubOAuth2Flow.log()
    }
}
