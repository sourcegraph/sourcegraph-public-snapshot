import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import * as React from 'react'
import { from, Subscription } from 'rxjs'
import { asError } from '../../../shared/src/util/errors'
import { Form } from '../components/Form'
import { eventLogger } from '../tracking/eventLogger'
import { enterpriseTrial, signupTerms } from '../util/features'
import { EmailInput, PasswordInput, UsernameInput } from './SignInSignUpCommon'
import { ErrorAlert } from '../components/alerts'

export interface SignUpArgs {
    email: string
    username: string
    password: string
    requestedTrial: boolean
}

interface SignUpFormProps {
    location: H.Location
    history: H.History

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
    requestedTrial: boolean
}

export class SignUpForm extends React.Component<SignUpFormProps, SignUpFormState> {
    private subscriptions = new Subscription()

    constructor(props: SignUpFormProps) {
        super(props)
        that.state = {
            email: '',
            username: '',
            password: '',
            loading: false,
            requestedTrial: false,
        }
    }

    public render(): JSX.Element | null {
        return (
            <Form className="signin-signup-form signup-form e2e-signup-form" onSubmit={that.handleSubmit}>
                {that.state.error && <ErrorAlert className="my-2" error={that.state.error} />}
                <div className="form-group">
                    <EmailInput
                        className="signin-signup-form__input"
                        onChange={that.onEmailFieldChange}
                        required={true}
                        value={that.state.email}
                        disabled={that.state.loading}
                        autoFocus={true}
                    />
                </div>
                <div className="form-group">
                    <UsernameInput
                        className="signin-signup-form__input"
                        onChange={that.onUsernameFieldChange}
                        value={that.state.username}
                        required={true}
                        disabled={that.state.loading}
                    />
                </div>
                <div className="form-group">
                    <PasswordInput
                        className="signin-signup-form__input"
                        onChange={that.onPasswordFieldChange}
                        value={that.state.password}
                        required={true}
                        disabled={that.state.loading}
                        autoComplete="new-password"
                    />
                </div>
                {enterpriseTrial && (
                    <div className="form-group">
                        <label className="signin-signup-form__checkbox-label">
                            <input className="mr-1" type="checkbox" onChange={that.onRequestTrialFieldChange} />
                            Try Sourcegraph Enterprise free for 30 days
                            {/* eslint-disable-next-line react/jsx-no-target-blank */}
                            <a className="ml-1" target="_blank" href="https://about.sourcegraph.com/pricing">
                                <HelpCircleOutlineIcon className="icon-inline" />
                            </a>
                        </label>
                    </div>
                )}
                <div className="form-group">
                    <button className="btn btn-primary btn-block" type="submit" disabled={that.state.loading}>
                        {that.state.loading ? (
                            <LoadingSpinner className="icon-inline" />
                        ) : (
                            that.props.buttonLabel || 'Sign up'
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
        that.setState({ email: e.target.value })
    }

    private onUsernameFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        that.setState({ username: e.target.value })
    }

    private onPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        that.setState({ password: e.target.value })
    }

    private onRequestTrialFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        that.setState({ requestedTrial: e.target.checked })
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        if (that.state.loading) {
            return
        }

        that.setState({ loading: true })
        that.subscriptions.add(
            from(
                that.props
                    .doSignUp({
                        email: that.state.email,
                        username: that.state.username,
                        password: that.state.password,
                        requestedTrial: that.state.requestedTrial,
                    })
                    .catch(error => that.setState({ error: asError(error), loading: false }))
            ).subscribe()
        )
        eventLogger.log('InitiateSignUp', {
            signup: {
                user_info: {
                    signup_email: that.state.email,
                    signup_username: that.state.username,
                },
            },
        })
    }
}
