import KeyIcon from '@sourcegraph/icons/lib/Key'
import { WebAuth } from 'auth0-js'
import * as React from 'react'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { sourcegraphContext } from '../../util/sourcegraphContext'

const webAuth = new WebAuth({
    domain: sourcegraphContext.auth0Domain,
    clientID: sourcegraphContext.auth0ClientID,
    redirectUri: `${sourcegraphContext.appURL}/-/auth0/sign-in`,
    responseType: 'code'
})

interface State {
    email: string
    error: string
    didReset: boolean
}

class PasswordResetForm extends React.Component<{}, State> {
    public state: State = {
        email: '',
        error: '',
        didReset: false
    }

    public render(): JSX.Element | null {
        if (this.state.didReset) {
            return <p className='password-reset-page__reset-confirm'>Password reset email sent.</p>
        }
        return (
            <form className='password-reset-page__form' onSubmit={this.handleSubmit}>
                {this.state.error !== '' &&
                    <p className='password-reset-page__error'>{this.state.error}</p>
                }
                <div>Enter your email address and we will send you a link to reset your password.</div>
                <input className='ui-text-box' onChange={this.onEmailFieldChange} value={this.state.email} type='email' placeholder='Email' />
                <button className='ui-button password-reset-page__submit' type='submit'>Reset Password</button>
            </form>
        )
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ email: e.target.value })
    }

    private handleSubmit = event => {
        event.preventDefault()

        webAuth.changePassword({
            connection: 'Sourcegraph',
            email: this.state.email
        }, (err, authResult) => {
            if (err) {
                console.error('auth error: ', err)
                this.setState({ error: (err as any).description })
            } else {
                this.setState({ didReset: true })
            }
        })
    }
}

/**
 * A landing page for the user request a password reset.
 */
export class PasswordResetPage extends React.Component {

    public render(): JSX.Element | null {
        return (
            <div className='password-reset-page'>
                <PageTitle title='reset password' />
                <HeroPage icon={KeyIcon} title='Sourcegraph' subtitle='Sign in or sign up to create an account' cta={<PasswordResetForm />} />
            </div>
        )
    }
}
