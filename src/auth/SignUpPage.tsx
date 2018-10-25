import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { upperFirst } from 'lodash'
import UserIcon from 'mdi-react/UserIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Redirect } from 'react-router-dom'
import { from, Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { Form } from '../components/Form'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { asError } from '../util/errors'
import { signupTerms } from '../util/features'
import { EmailInput, getReturnTo, PasswordInput, UsernameInput } from './SignInSignUpCommon'

export interface SignUpArgs {
    email: string
    username: string
    password: string
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
            <Form className="signin-signup-form signup-form" onSubmit={this.handleSubmit}>
                {this.state.error && (
                    <div className="alert alert-danger my-2">Error: {upperFirst(this.state.error.message)}</div>
                )}
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
                {signupTerms && (
                    <small className="form-text text-muted">
                        By signing up, you agree to our{' '}
                        <a href="https://about.sourcegraph.com/terms" target="_blank">
                            Terms of Service
                        </a>{' '}
                        and{' '}
                        <a href="https://about.sourcegraph.com/privacy" target="_blank">
                            Privacy Policy
                        </a>.
                    </small>
                )}
            </Form>
        )
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ email: e.target.value })
    }

    private onUsernameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ username: e.target.value })
    }

    private onPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ password: e.target.value })
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
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
        eventLogger.log('InitiateSignUp', {
            signup: {
                user_info: {
                    signup_email: this.state.email,
                    signup_username: this.state.username,
                },
            },
        })
    }
}

interface SignUpPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
}

export class SignUpPage extends React.Component<SignUpPageProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SignUp', {}, false)
    }

    public render(): JSX.Element | null {
        if (this.props.authenticatedUser) {
            const returnTo = getReturnTo(this.props.location)
            return <Redirect to={returnTo} />
        }

        if (!window.context.allowSignup) {
            return <Redirect to="/sign-in" />
        }

        return (
            <div className="signin-signup-page sign-up-page">
                <PageTitle title="Sign up" />
                <HeroPage
                    icon={UserIcon}
                    title="Sign up for Sourcegraph"
                    cta={
                        <div>
                            <Link className="signin-signup-form__mode" to={`/sign-in?${this.props.location.search}`}>
                                Already have an account? Sign in.
                            </Link>
                            <SignUpForm {...this.props} doSignUp={this.doSignUp} />
                        </div>
                    }
                />
            </div>
        )
    }

    private doSignUp = (args: SignUpArgs): Promise<void> =>
        fetch('/-/sign-up', {
            credentials: 'same-origin',
            method: 'POST',
            headers: {
                ...window.context.xhrHeaders,
                Accept: 'application/json',
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(args),
        }).then(resp => {
            if (resp.status !== 200) {
                return resp.text().then(text => Promise.reject(new Error(text)))
            }
            window.location.replace(getReturnTo(this.props.location))
            return Promise.resolve()
        })
}
