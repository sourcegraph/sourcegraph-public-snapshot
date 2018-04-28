import KeyIcon from '@sourcegraph/icons/lib/Key'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Form } from '../components/Form'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { PasswordInput } from './SignInSignUpCommon'

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
                <Form className="password-reset-page__form" onSubmit={this.handleSubmitResetPassword}>
                    {this.state.error !== '' && (
                        <p className="password-reset-page__error">{upperFirst(this.state.error)}</p>
                    )}
                    <p>Enter your new password.</p>
                    <div className="form-group">
                        <PasswordInput
                            name="password"
                            onChange={this.onPasswordFieldChange}
                            value={this.state.password}
                            required={true}
                            autoComplete="new-password"
                        />
                    </div>
                    <button className="btn btn-primary btn-block" type="submit">
                        Reset Password
                    </button>
                </Form>
            )
        }

        // If `code` and `email` aren't provided, assume the user has not initiated the password reset
        // flow and display a form for the user to enter their email to receive a password reset email.
        if (this.state.didReset) {
            return <p className="password-reset-page__reset-confirm">Password reset email sent.</p>
        }
        return (
            <Form className="password-reset-page__form" onSubmit={this.handleSubmitResetPasswordInit}>
                {this.state.error !== '' && (
                    <p className="password-reset-page__error">{upperFirst(this.state.error)}</p>
                )}
                <p>Enter your email address and we will send you a link to reset your password.</p>
                <div className="form-group">
                    <input
                        className="form-control"
                        onChange={this.onEmailFieldChange}
                        value={this.state.email}
                        type="email"
                        name="email"
                        spellCheck={false}
                        placeholder="Email"
                        required={true}
                        autoComplete="email"
                    />
                </div>
                <button className="btn btn-primary btn-block" type="submit">
                    Reset Password
                </button>
            </Form>
        )
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ email: e.target.value })
    }

    private onPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ password: e.target.value })
    }

    private handleSubmitResetPasswordInit = (e: React.FormEvent<HTMLFormElement>) => {
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
                } else if (resp.status === 429) {
                    this.setState({ error: 'Too many password reset requests. Try again in a few minutes.' })
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
    public componentDidMount(): void {
        eventLogger.logViewEvent('PasswordReset')
    }

    public render(): JSX.Element | null {
        return (
            <div className="password-reset-page">
                <PageTitle title="Reset password" />
                <HeroPage icon={KeyIcon} title="Reset password" cta={<PasswordResetForm />} />
            </div>
        )
    }
}
