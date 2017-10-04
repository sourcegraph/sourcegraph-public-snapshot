import KeyIcon from '@sourcegraph/icons/lib/Key'
import Loader from '@sourcegraph/icons/lib/Loader'
import { Auth0Error, WebAuth } from 'auth0-js'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { events, viewEvents } from '../tracking/events'
import { sourcegraphContext } from '../util/sourcegraphContext'

interface LoginSignupFormProps {
    location: H.Location
}

interface LoginSignupFormState {
    mode: 'signin' | 'signup'
    email: string
    password: string
    errorDescription: string
    loading: boolean
}

class LoginSignupForm extends React.Component<LoginSignupFormProps, LoginSignupFormState> {
    public state: LoginSignupFormState = {
        mode: 'signin',
        email: '',
        password: '',
        errorDescription: '',
        loading: false
    }

    public render(): JSX.Element | null {
        return (
            <form className='login-signup-form' onSubmit={this.handleSubmit}>
                {this.state.errorDescription !== '' &&
                    <p className='login-signup-form__error'>{this.state.errorDescription}</p>
                }
                <div className='login-signup-form__modes'>
                    <span className={`login-signup-form__mode${this.state.mode === 'signin' ? '--active' : ''}`} onClick={this.setModeSignIn}>Sign in</span>
                    <span className='login-signup-form__mode-divider'>|</span>
                    <span className={`login-signup-form__mode${this.state.mode === 'signup' ? '--active' : ''}`} onClick={this.setModeSignUp}>Sign up</span>
                </div>
                <div className='form-group'>
                    <input
                        className='ui-text-box login-signup-form__input'
                        onChange={this.onEmailFieldChange}
                        value={this.state.email}
                        type='email'
                        placeholder='Email'
                        disabled={this.state.loading}
                    />
                </div>
                <div className='form-group'>
                    <input
                        className='ui-text-box login-signup-form__input'
                        onChange={this.onPasswordFieldChange}
                        value={this.state.password}
                        type='password'
                        placeholder='Password'
                        disabled={this.state.loading}
                    />
                    <small className='form-text'><Link to='/password-reset'>Forgot password?</Link></small>
                </div>
                <div className='form-group'>
                    <button className='btn btn-primary btn-block' type='submit' disabled={this.state.loading}>
                        {this.state.mode === 'signin' ? 'Sign In' : 'Sign Up'}
                    </button>
                </div>
                <small className='form-text'>Existing users who signed in via GitHub: please sign up for a Sourcegraph account.</small>
                {this.state.loading && <div className='login-signup-form__loader'><Loader className='icon-inline'/></div>}
            </form>
        )
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ email: e.target.value })
    }

    private onPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ password: e.target.value })
    }

    private setModeSignIn = () => {
        this.setState({ mode: 'signin' })
    }

    private setModeSignUp = () => {
        this.setState({ mode: 'signup' })
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        const redirect = new URL(`${sourcegraphContext.appURL}/-/auth0/sign-in`)
        const returnTo = new URLSearchParams(this.props.location.search).get('return-to')
        if (returnTo) {
            redirect.searchParams.set('return-to', returnTo)
        }

        const webAuth = new WebAuth({
            domain: sourcegraphContext.auth0Domain,
            clientID: sourcegraphContext.auth0ClientID,
            redirectUri: redirect.href,
            responseType: 'code'
        })

        event.preventDefault()
        if (this.state.loading) {
            return
        }
        this.setState({ loading: true })
        const authCallback = (err: Auth0Error) => {
            this.setState({ loading: false })
            if (err) {
                console.error('auth error: ', err)
                this.setState({ errorDescription: err.description || 'Unknown Error' })
            }
        }
        switch (this.state.mode) {
            case 'signin':
                events.InitiateSignIn.log({
                    signup: {
                        user_info: {
                            signup_email: this.state.email
                        }
                    }
                })
                webAuth.redirect.loginWithCredentials({
                    connection: 'Sourcegraph',
                    responseType: 'code',
                    username: this.state.email,
                    password: this.state.password
                }, authCallback)
                break
            case 'signup':
                events.InitiateSignUp.log({
                    signup: {
                        user_info: {
                            signup_email: this.state.email
                        }
                    }
                })
                webAuth.redirect.signupAndLogin({
                    connection: 'Sourcegraph',
                    responseType: 'code',
                    email: this.state.email,
                    password: this.state.password
                }, authCallback)
                break
        }
    }
}

interface SignInPageProps {
    location: H.Location
}

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class SignInPage extends React.Component<SignInPageProps> {

    public componentDidMount(): void {
        viewEvents.SignIn.log()
    }

    public render(): JSX.Element | null {
        if (sourcegraphContext.user) {
            return <Redirect to='/search' />
        }

        return (
            <div className='sign-in-page'>
                <PageTitle title='Sign in or sign up' />
                <HeroPage icon={KeyIcon} title='Welcome to Sourcegraph' subtitle='Sign in or sign up to create an account' cta={<LoginSignupForm {...this.props} />} />
            </div>
        )
    }
}
