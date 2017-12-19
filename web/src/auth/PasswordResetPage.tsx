import KeyIcon from '@sourcegraph/icons/lib/Key'
import { WebAuth } from 'auth0-js'
import * as React from 'react'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { VALID_PASSWORD_REGEXP } from '../settings/validation'

const webAuth = new WebAuth({
    domain: window.context.auth0Domain,
    clientID: window.context.auth0ClientID,
    redirectUri: `${window.context.appURL}/-/auth0/sign-in`,
    responseType: 'code',
})

interface State {
    email: string
    error: string
    didReset: boolean

    password: string
}

class PasswordResetForm extends React.Component<{}, State> {
    public state: State = {
        email: '',
        error: '',
        didReset: false,

        password: '',
    }

    public render(): JSX.Element | null {
        const searchParams = new URLSearchParams(window.location.search)
        const code = searchParams.get('code')
        const email = searchParams.get('email')
        if (code && email) {
            // If `code` and `email` are provided in the URL, then display a form that reset a user's password
            // on submission if the code matches the email.
            return (
                <form className="password-reset-page__form" onSubmit={this.handleSubmitResetPassword}>
                    {this.state.error !== '' && <p className="password-reset-page__error">{this.state.error}</p>}
                    <p>Enter your new password.</p>
                    <div className="form-group">
                        <input
                            className="form-control"
                            onChange={this.onPasswordFieldChange}
                            value={this.state.password}
                            type="password"
                            spellCheck={false}
                            placeholder="Password"
                            pattern={VALID_PASSWORD_REGEXP.toString().slice(1, -1)}
                            required={true}
                        />
                    </div>
                    <button className="btn btn-primary btn-block" type="submit">
                        Reset Password
                    </button>
                </form>
            )
        }

        // If `code` and `email` aren't provided, assume the user has not initiated the password reset
        // flow and display a form for the user to enter their email to receive a password reset email.
        if (this.state.didReset) {
            return <p className="password-reset-page__reset-confirm">Password reset email sent.</p>
        }
        return (
            <form className="password-reset-page__form" onSubmit={this.handleSubmitResetPasswordInit}>
                {this.state.error !== '' && <p className="password-reset-page__error">{this.state.error}</p>}
                <p>Enter your email address and we will send you a link to reset your password.</p>
                <div className="form-group">
                    <input
                        className="form-control"
                        onChange={this.onEmailFieldChange}
                        value={this.state.email}
                        type="email"
                        spellCheck={false}
                        placeholder="Email"
                        required={true}
                    />
                </div>
                <button className="btn btn-primary btn-block" type="submit">
                    Reset Password
                </button>
            </form>
        )
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ email: e.target.value })
    }

    private onPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ password: e.target.value })
    }

    private handleSubmitResetPasswordInit = (e: React.FormEvent<HTMLFormElement>) => {
        if (window.context.useAuth0) {
            // Legacy Auth0 path
            this.handleSubmitResetPasswordInitAuth0(e)
        } else {
            this.handleSubmitResetPasswordInitNative(e)
        }
    }

    private handleSubmitResetPasswordInitAuth0(e: React.FormEvent<HTMLFormElement>): void {
        e.preventDefault()

        webAuth.changePassword(
            {
                connection: 'Sourcegraph',
                email: this.state.email,
            },
            (err, authResult) => {
                if (err) {
                    console.error('auth error: ', err)
                    this.setState({ error: (err as any).description })
                } else {
                    this.setState({ didReset: true })
                }
            }
        )
    }

    private handleSubmitResetPasswordInitNative(e: React.FormEvent<HTMLFormElement>): void {
        e.preventDefault()
        fetch('/-/reset-password-init', {
            credentials: 'same-origin',
            method: 'POST',
            headers: {
                ...window.context.xhrHeaders,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email: this.state.email,
            }),
        })
            .then(resp => {
                if (resp.status === 200) {
                    this.setState({ didReset: true })
                } else {
                    this.setState({ error: 'Could not reset password: Status code ' + resp.status })
                }
            })
            .catch(err => {
                this.setState({ error: 'Could not reset password: ' + (err && err.message) || err || 'Unknown Error' })
            })
    }

    private handleSubmitResetPassword = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()

        const searchParams = new URLSearchParams(window.location.search)
        const code = searchParams.get('code')
        const email = searchParams.get('email')

        fetch('/-/reset-password', {
            credentials: 'same-origin',
            method: 'POST',
            headers: {
                ...window.context.xhrHeaders,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, code, password: this.state.password }),
        })
            .then(resp => {
                if (resp.status === 200) {
                    window.location.replace('/sign-in')
                } else if (resp.status === 401) {
                    this.setState({ error: 'Password reset code was invalid or expired.' })
                } else {
                    this.setState({ error: 'Password reset failed. Status code: ' + resp.status })
                }
            })
            .catch(err => {
                this.setState({ error: 'Password reset failed: ' + (err && err.message) || err || 'Unknown Error' })
            })
    }
}

/**
 * A landing page for the user request a password reset.
 */
export class PasswordResetPage extends React.Component {
    public render(): JSX.Element | null {
        return (
            <div className="password-reset-page">
                <PageTitle title="Reset password" />
                <HeroPage
                    icon={KeyIcon}
                    title="Sourcegraph"
                    subtitle="Sign in or sign up to create an account"
                    cta={<PasswordResetForm />}
                />
            </div>
        )
    }
}
