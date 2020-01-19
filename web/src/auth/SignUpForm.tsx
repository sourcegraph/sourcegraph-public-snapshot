import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { from, Subscription } from 'rxjs'
import { asError } from '../../../shared/src/util/errors'
import { Form } from '../components/Form'
import { eventLogger } from '../tracking/eventLogger'
import { signupTerms } from '../util/features'
import { EmailInput, PasswordInput, UsernameInput } from './SignInSignUpCommon'
import { ErrorAlert } from '../components/alerts'

export interface SignUpArgs {
    email: string
    username: string
    password: string
}

interface SignUpFormProps {
    /** Called to perform the signup on the server. */
    doSignUp: (args: SignUpArgs) => Promise<void>

    buttonLabel?: string
}

interface SignUpFormState {
    email: string
    username: string
    password: string
    error?: Error
    loading: boolean
}

export class SignUpForm extends React.Component<SignUpFormProps, SignUpFormState> {
    private subscriptions = new Subscription()

    constructor(props: SignUpFormProps) {
        super(props)
        this.state = {
            email: '',
            username: '',
            password: '',
            loading: false,
        }
    }

    public render(): JSX.Element | null {
        return (
            <Form className="signin-signup-form signup-form e2e-signup-form" onSubmit={this.handleSubmit}>
                {this.state.error && <ErrorAlert className="my-2" error={this.state.error} />}
                <div className="form-group">
                    <EmailInput
                        className="signin-signup-form__input"
                        onChange={this.onEmailFieldChange}
                        required={true}
                        value={this.state.email}
                        disabled={this.state.loading}
                        autoFocus={true}
                    />
                </div>
                <div className="form-group">
                    <UsernameInput
                        className="signin-signup-form__input"
                        onChange={this.onUsernameFieldChange}
                        value={this.state.username}
                        required={true}
                        disabled={this.state.loading}
                    />
                </div>
                <div className="form-group">
                    <PasswordInput
                        className="signin-signup-form__input"
                        onChange={this.onPasswordFieldChange}
                        value={this.state.password}
                        required={true}
                        disabled={this.state.loading}
                        autoComplete="new-password"
                    />
                </div>
                <div className="form-group">
                    <button className="btn btn-primary btn-block" type="submit" disabled={this.state.loading}>
                        {this.state.loading ? (
                            <LoadingSpinner className="icon-inline" />
                        ) : (
                            this.props.buttonLabel || 'Sign up'
                        )}
                    </button>
                </div>
                {window.context.sourcegraphDotComMode && (
                    <p>
                        Create a public account to search/navigate open-source code and manage Sourcegraph
                        subscriptions.
                    </p>
                )}
                {signupTerms && (
                    <small className="form-text text-muted">
                        By signing up, you agree to our
                        {/* eslint-disable-next-line react/jsx-no-target-blank */}
                        <a href="https://about.sourcegraph.com/terms" target="_blank">
                            Terms of Service
                        </a>{' '}
                        and {/* eslint-disable-next-line react/jsx-no-target-blank */}
                        <a href="https://about.sourcegraph.com/privacy" target="_blank">
                            Privacy Policy
                        </a>
                        .
                    </small>
                )}
            </Form>
        )
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ email: e.target.value })
    }

    private onUsernameFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ username: e.target.value })
    }

    private onPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ password: e.target.value })
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        if (this.state.loading) {
            return
        }

        this.setState({ loading: true })
        this.subscriptions.add(
            from(
                this.props
                    .doSignUp({
                        email: this.state.email,
                        username: this.state.username,
                        password: this.state.password,
                    })
                    .catch(error => this.setState({ error: asError(error), loading: false }))
            ).subscribe()
        )
        eventLogger.log('InitiateSignUp')
    }
}
