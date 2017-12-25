import Loader from '@sourcegraph/icons/lib/Loader'
import UserIcon from '@sourcegraph/icons/lib/User'
import { WebAuth } from 'auth0-js'
import * as H from 'history'
import { Base64 } from 'js-base64'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Redirect } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { VALID_USERNAME_REGEXP } from '../settings/validation'
import { eventLogger } from '../tracking/eventLogger'
import { signupTerms } from '../util/features'
import { isUsernameAvailable } from './backend'
import { EmailInput, getReturnTo, PasswordInput } from './SignInSignUpCommon'

interface SignUpFormProps {
    location: H.Location
    history: H.History
    prefilledEmail?: string
}

interface SignUpFormState {
    email: string
    username: string
    password: string
    errorDescription: string
    loading: boolean
}

class SignUpForm extends React.Component<SignUpFormProps, SignUpFormState> {
    private subscriptions = new Subscription()

    constructor(props: SignUpFormProps) {
        super(props)
        this.state = {
            email: props.prefilledEmail || '',
            username: '',
            password: '',
            errorDescription: '',
            loading: false,
        }
    }

    public render(): JSX.Element | null {
        return (
            <form className="signin-signup-form signup-form" onSubmit={this.handleSubmit}>
                {this.state.errorDescription !== '' && (
                    <p className="signin-signup-form__error">{this.state.errorDescription}</p>
                )}
                <Link className="signin-signup-form__mode" to={`/sign-in?${this.props.location.search}`}>
                    Already have an account? Sign in.
                </Link>
                <div className="form-group">
                    <EmailInput
                        className="signin-signup-form__input"
                        onChange={this.onEmailFieldChange}
                        required={true}
                        value={this.state.email}
                        disabled={this.state.loading || Boolean(this.props.prefilledEmail)}
                    />
                </div>
                <div className="form-group">
                    <input
                        className="form-control signin-signup-form__input"
                        onChange={this.onUsernameFieldChange}
                        value={this.state.username}
                        type="text"
                        required={true}
                        placeholder="Username"
                        pattern={VALID_USERNAME_REGEXP.toString().slice(1, -1)}
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
                    />
                </div>
                <div className="form-group">
                    <button className="btn btn-primary btn-block" type="submit" disabled={this.state.loading}>
                        Sign up
                    </button>
                </div>
                {signupTerms && (
                    <small className="form-text signup-form__terms">
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
                    <div className="signin-signup-form__loader">
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

    private onPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ password: e.target.value })
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        if (window.context.useAuth0) {
            // Legacy Auth0 path
            this.handleSubmitAuth0(event)
        } else {
            this.handleSubmitNative(event)
        }
    }

    private handleSubmitAuth0(event: React.FormEvent<HTMLFormElement>): void {
        const redirect = new URL(`${window.context.appURL}/-/auth0/sign-in`)
        const searchParams = new URLSearchParams(this.props.location.search)
        const returnTo = getReturnTo(this.props.location)
        if (returnTo) {
            redirect.searchParams.set('returnTo', returnTo)
        }
        redirect.searchParams.set('username', this.state.username)
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
                        user_metadata: { name: this.state.username },
                    },
                    authCallback
                )
            })
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

    private handleSubmitNative(event: React.FormEvent<HTMLFormElement>): void {
        event.preventDefault()
        if (this.state.loading) {
            return
        }

        this.setState({ loading: true })
        this.subscriptions.add(
            isUsernameAvailable(this.state.username).subscribe(availability => {
                if (!availability) {
                    this.setState({
                        errorDescription: 'The username you selected is already taken, please try again.',
                        loading: false,
                    })
                    return
                } else {
                    fetch('/-/sign-up', {
                        credentials: 'same-origin',
                        method: 'POST',
                        headers: {
                            ...window.context.xhrHeaders,
                            Accept: 'application/json',
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            email: this.state.email,
                            password: this.state.password,
                            username: this.state.username,
                        }),
                    }).then(
                        resp => {
                            if (resp.status === 200) {
                                const returnTo = getReturnTo(this.props.location)
                                if (returnTo) {
                                    window.location.replace(returnTo)
                                } else {
                                    window.location.replace('/search')
                                }
                            } else {
                                this.setState({
                                    errorDescription: 'Could not create user',
                                    loading: false,
                                })
                            }
                        },
                        err => {
                            this.setState({
                                errorDescription: 'Could not create user',
                                loading: false,
                            })
                        }
                    )
                }
            })
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
    user: GQL.IUser | null
}

interface SignUpPageState {
    prefilledEmail?: string
}

export class SignUpPage extends React.Component<SignUpPageProps, SignUpPageState> {
    constructor(props: SignUpPageProps) {
        super(props)
        this.state = {
            prefilledEmail: this.getPrefilledEmail(props),
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('SignUp')
    }

    public componentWillReceiveProps(nextProps: SignUpPageProps): void {
        this.setState({ prefilledEmail: this.getPrefilledEmail(nextProps) })
    }

    public render(): JSX.Element | null {
        if (this.props.user) {
            const returnTo = getReturnTo(this.props.location)
            if (returnTo) {
                return <Redirect to={returnTo} />
            }
            return <Redirect to="/search" />
        }

        return (
            <div className="signin-signup-page sign-up-page">
                <PageTitle title="Sign up" />
                <HeroPage
                    icon={UserIcon}
                    title="Sign up for Sourcegraph"
                    cta={<SignUpForm {...this.props} prefilledEmail={this.state.prefilledEmail} />}
                />
            </div>
        )
    }

    private getPrefilledEmail(props: SignUpPageProps): string | undefined {
        const searchParams = new URLSearchParams(props.location.search)
        let prefilledEmail: string | undefined
        if (searchParams.get('token')) {
            const tokenPayload = JSON.parse(Base64.decode(searchParams.get('token')!.split('.')[1]))
            prefilledEmail = tokenPayload.email
        }
        return prefilledEmail
    }
}
