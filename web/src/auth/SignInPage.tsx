import KeyIcon from '@sourcegraph/icons/lib/Key'
import Loader from '@sourcegraph/icons/lib/Loader'
import { WebAuth } from 'auth0-js'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { userForgotPassword } from '../util/features'
import { EmailInput, getReturnTo, PasswordInput } from './SignInSignUpCommon'

interface SignInFormProps {
    location: H.Location
    history: H.History
}

interface SignInFormState {
    email: string
    password: string
    errorDescription: string
    loading: boolean
}

class SignInForm extends React.Component<SignInFormProps, SignInFormState> {
    constructor(props: SignInFormProps) {
        super(props)
        this.state = {
            email: '',
            password: '',
            errorDescription: '',
            loading: false,
        }
    }

    public render(): JSX.Element | null {
        return (
            <form className="signin-signup-form signin-form" onSubmit={this.handleSubmit}>
                {this.state.errorDescription !== '' && (
                    <p className="signin-signup-form__error">{this.state.errorDescription}</p>
                )}
                {window.context.site['auth.allowSignup'] && (
                    <Link className="signin-signup-form__mode" to={`/sign-up${this.props.location.search}`}>
                        Don't have an account? Sign up.
                    </Link>
                )}
                <div className="form-group">
                    <EmailInput
                        className="signin-signup-form__input"
                        onChange={this.onEmailFieldChange}
                        required={true}
                        value={this.state.email}
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
                    {userForgotPassword && (
                        <small className="form-text">
                            <Link to="/password-reset">Forgot password?</Link>
                        </small>
                    )}
                </div>
                <div className="form-group">
                    <button className="btn btn-primary btn-block" type="submit" disabled={this.state.loading}>
                        Sign in
                    </button>
                </div>
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
    }

    private handleSubmitNative(event: React.FormEvent<HTMLFormElement>): void {
        event.preventDefault()
        if (this.state.loading) {
            return
        }

        this.setState({ loading: true })
        eventLogger.log('InitiateSignIn')
        fetch('/-/sign-in-2', {
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
            }),
        })
            .then(resp => {
                if (resp.status === 200) {
                    const returnTo = getReturnTo(this.props.location)
                    if (returnTo) {
                        window.location.replace(returnTo)
                    } else {
                        window.location.replace('/search')
                    }
                } else if (resp.status === 401) {
                    throw new Error('User or password was incorrect')
                } else {
                    throw new Error('Unknown Error')
                }
            })
            .catch(err => {
                console.error('auth error: ', err)
                this.setState({ loading: false, errorDescription: (err && err.message) || 'Unknown Error' })
            })
    }
}

interface SignInPageProps {
    location: H.Location
    history: H.History
    user: GQL.IUser | null
}

export class SignInPage extends React.Component<SignInPageProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SignIn')
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
            <div className="signin-signup-page sign-in-page">
                <PageTitle title="Sign in" />
                <HeroPage icon={KeyIcon} title="Sign into Sourcegraph" cta={<SignInForm {...this.props} />} />
            </div>
        )
    }
}
