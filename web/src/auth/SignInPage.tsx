import KeyIcon from '@sourcegraph/icons/lib/Key'
import Loader from '@sourcegraph/icons/lib/Loader'
import { WebAuth } from 'auth0-js'
import * as H from 'history'
import { Base64 } from 'js-base64'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Redirect } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { VALID_PASSWORD_REGEXP, VALID_USERNAME_REGEXP } from '../settings/validation'
import { eventLogger } from '../tracking/eventLogger'
import { isUsernameAvailable } from './backend'

interface LoginSignupFormProps {
    location: H.Location
    history: H.History
    mode: 'signin' | 'signup'
    prefilledEmail?: string
}

interface LoginSignupFormState {
    email: string
    username: string
    displayName: string
    password: string
    errorDescription: string
    loading: boolean
}

class LoginSignupForm extends React.Component<LoginSignupFormProps, LoginSignupFormState> {
    private subscriptions = new Subscription()

    constructor(props: LoginSignupFormProps) {
        super(props)
        this.state = {
            email: props.prefilledEmail || '',
            username: '',
            displayName: '',
            password: '',
            errorDescription: '',
            loading: false,
        }
    }

    public render(): JSX.Element | null {
        return (
            <form className="login-signup-form" onSubmit={this.handleSubmit}>
                {this.state.errorDescription !== '' && (
                    <p className="login-signup-form__error">{this.state.errorDescription}</p>
                )}
                <div className="login-signup-form__modes">
                    <span
                        className={`login-signup-form__mode${this.props.mode === 'signin' ? '--active' : ''}`}
                        onClick={this.setModeSignIn}
                    >
                        Sign in
                    </span>
                    <span className="login-signup-form__mode-divider">|</span>
                    <span
                        className={`login-signup-form__mode${this.props.mode === 'signup' ? '--active' : ''}`}
                        onClick={this.setModeSignUp}
                    >
                        Sign up
                    </span>
                </div>
                <div className="form-group">
                    <input
                        className="form-control login-signup-form__input"
                        onChange={this.onEmailFieldChange}
                        value={this.state.email}
                        type="email"
                        placeholder="Email"
                        required={true}
                        disabled={this.state.loading || Boolean(this.props.prefilledEmail)}
                        spellCheck={false}
                    />
                </div>
                {this.props.mode === 'signup' && (
                    <div className="form-group">
                        <input
                            className="form-control login-signup-form__input"
                            onChange={this.onUsernameFieldChange}
                            value={this.state.username}
                            type="text"
                            required={true}
                            placeholder="Username"
                            pattern={VALID_USERNAME_REGEXP.toString().slice(1, -1)}
                            disabled={this.state.loading}
                        />
                    </div>
                )}
                <div className="form-group">
                    <input
                        className="form-control login-signup-form__input"
                        onChange={this.onPasswordFieldChange}
                        value={this.state.password}
                        required={true}
                        type="password"
                        placeholder="Password"
                        pattern={VALID_PASSWORD_REGEXP.toString().slice(1, -1)}
                        disabled={this.state.loading}
                    />
                    {this.props.mode === 'signin' && (
                        <small className="form-text">
                            <Link to="/password-reset">Forgot password?</Link>
                        </small>
                    )}
                </div>
                {this.props.mode === 'signup' && (
                    <div className="form-group">
                        <input
                            className="form-control login-signup-form__input"
                            onChange={this.onDisplayNameFieldChange}
                            value={this.state.displayName}
                            type="text"
                            placeholder="Display name (optional)"
                            disabled={this.state.loading}
                        />
                    </div>
                )}
                <div className="form-group">
                    <button className="btn btn-primary btn-block" type="submit" disabled={this.state.loading}>
                        {this.props.mode === 'signin' ? 'Sign In' : 'Sign Up'}
                    </button>
                </div>
                <small className="form-text">
                    Existing users who signed in via GitHub: please sign up for a Sourcegraph account.
                </small>
                {this.props.mode === 'signup' && (
                    <small className="form-text sign-in-page__terms">
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
                {this.state.loading && (
                    <div className="login-signup-form__loader">
                        <Loader className="icon-inline" />
                    </div>
                )}
            </form>
        )
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ email: e.target.value })
    }

    private onUsernameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ username: e.target.value })
    }

    private onDisplayNameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ displayName: e.target.value })
    }

    private onPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ password: e.target.value })
    }

    private setModeSignIn = () => {
        this.props.history.push(`/sign-in` + this.props.location.search)
    }

    private setModeSignUp = () => {
        this.props.history.push(`/sign-up` + this.props.location.search)
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        const redirect = new URL(`${window.context.appURL}/-/auth0/sign-in`)
        const searchParams = new URLSearchParams(this.props.location.search)
        const returnTo = searchParams.get('returnTo')

        if (returnTo) {
            // 🚨 SECURITY: important that we do not allow redirects to
            // arbitrary hosts here.
            const newURL = new URL(returnTo, window.location.href)
            redirect.searchParams.set('returnTo', window.context.appURL + newURL.pathname + newURL.search + newURL.hash)
        }
        if (this.props.mode === 'signup') {
            redirect.searchParams.set('username', this.state.username)
            redirect.searchParams.set('displayName', this.state.displayName || this.state.username)
        }
        const token = searchParams.get('token')
        if (token) {
            redirect.searchParams.set('token', token)
        }

        const webAuth = new WebAuth({
            domain: window.context.auth0Domain,
            clientID: window.context.auth0ClientID,
            redirectUri: redirect.href,
            responseType: 'code',
        })

        event.preventDefault()
        if (this.state.loading) {
            return
        }
        this.setState({ loading: true })
        const authCallback = (err: any) => {
            this.setState({ loading: false })
            if (err) {
                console.error('auth error: ', err)
                this.setState({ errorDescription: err.description || 'Unknown Error' })
            }
        }
        switch (this.props.mode) {
            case 'signin':
                eventLogger.log('InitiateSignIn')
                webAuth.redirect.loginWithCredentials(
                    {
                        connection: 'Sourcegraph',
                        responseType: 'code',
                        username: this.state.email,
                        password: this.state.password,
                    },
                    authCallback
                )
                break
            case 'signup':
                this.subscriptions.add(
                    isUsernameAvailable(this.state.username).subscribe(availability => {
                        if (!availability) {
                            this.setState({
                                errorDescription: 'The username you selected is already taken, please try again.',
                                loading: false,
                            })
                            return
                        }
                        webAuth.redirect.signupAndLogin(
                            {
                                connection: 'Sourcegraph',
                                responseType: 'code',
                                email: this.state.email,
                                password: this.state.password,
                                // Setting user_metdata is a "nice-to-have" but doesn't correctly update the
                                // user's name in Auth0. That's not an issue per-se, see more at
                                // https://github.com/auth0/auth0.js/issues/70.
                                user_metadata: { name: this.state.displayName || this.state.username },
                            },
                            authCallback
                        )
                    })
                )
                eventLogger.log('InitiateSignUp', {
                    signup: {
                        user_info: {
                            signup_email: this.state.email,
                            signup_display_name: this.state.displayName,
                            signup_username: this.state.username,
                        },
                    },
                })
        }
    }
}

interface SignInPageProps {
    location: H.Location
    history: H.History
}

interface SignInPageState {
    prefilledEmail?: string
}

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class SignInPage extends React.Component<SignInPageProps, SignInPageState> {
    constructor(props: SignInPageProps) {
        super(props)
        this.state = {
            prefilledEmail: this.getPrefilledEmail(props),
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('SignIn')
    }

    public componentWillReceiveProps(nextProps: SignInPageProps): void {
        this.setState({ prefilledEmail: this.getPrefilledEmail(nextProps) })
    }

    public render(): JSX.Element | null {
        if (window.context.user) {
            const searchParams = new URLSearchParams(this.props.location.search)
            const returnTo = searchParams.get('returnTo')
            if (returnTo) {
                // 🚨 SECURITY: important that we do not allow redirects to
                // arbitrary hosts here.
                const newURL = new URL(returnTo, window.location.href)
                return <Redirect to={newURL.pathname + newURL.search + newURL.hash} />
            }
            return <Redirect to="/search" />
        }

        return (
            <div className="sign-in-page">
                <PageTitle title={this.props.location.pathname === '/sign-in' ? 'Sign in' : 'Sign up'} />
                <HeroPage
                    icon={KeyIcon}
                    title="Welcome to Sourcegraph"
                    cta={
                        <LoginSignupForm
                            {...this.props}
                            mode={this.props.location.pathname === '/sign-in' ? 'signin' : 'signup'}
                            prefilledEmail={this.state.prefilledEmail}
                        />
                    }
                />
            </div>
        )
    }

    private getPrefilledEmail(props: SignInPageProps): string | undefined {
        const searchParams = new URLSearchParams(props.location.search)
        let prefilledEmail: string | undefined
        if (searchParams.get('token')) {
            const tokenPayload = JSON.parse(Base64.decode(searchParams.get('token')!.split('.')[1]))
            prefilledEmail = tokenPayload.email
        }
        return prefilledEmail
    }
}
