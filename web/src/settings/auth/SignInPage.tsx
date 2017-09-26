import KeyIcon from '@sourcegraph/icons/lib/Key'
import { WebAuth } from 'auth0-js'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { events } from '../../tracking/events'
import { sourcegraphContext } from '../../util/sourcegraphContext'

const webAuth = new WebAuth({
    domain: sourcegraphContext.auth0Domain,
    clientID: sourcegraphContext.auth0ClientID,
    redirectUri: `${sourcegraphContext.appURL}/-/auth0/sign-in`,
    responseType: 'code'
})

interface LoginSignupFormProps {}

interface LoginSignupFormState {
    mode: 'signin' | 'signup'
    email: string
    password: string
    errorDescription: string
}

class LoginSignupForm extends React.Component<LoginSignupFormProps, LoginSignupFormState> {
    public state: LoginSignupFormState = {
        mode: 'signin',
        email: '',
        password: '',
        errorDescription: ''
    }

    public render(): JSX.Element | null {
        return (
            <form className='sign-in-page__form' onSubmit={this.handleSubmit}>
                {this.state.errorDescription !== '' &&
                    <p className='sign-in-page__error'>{this.state.errorDescription}</p>
                }
                <div className='sign-in-page__modes'>
                    <span className={`sign-in-page__mode${this.state.mode === 'signin' ? '--active' : ''}`} onClick={this.setModeSignIn}>Sign in</span>
                    <span className='sign-in-page__mode-divider'>|</span>
                    <span className={`sign-in-page__mode${this.state.mode === 'signup' ? '--active' : ''}`} onClick={this.setModeSignUp}>Sign up</span>
                </div>
                <input className='ui-text-box sign-in-page__input' onChange={this.onEmailFieldChange} value={this.state.email} type='email' placeholder='Email' />
                <input className='ui-text-box sign-in-page__input' onChange={this.onPasswordFieldChange} value={this.state.password} type='password' placeholder='Password' />
                <Link to='/password-reset' className='sign-in-page__pass-reset-link'><small>Forgot password?</small></Link>
                <div className='form-group sign-in-page__submit-button'>
                    <button className='btn btn-primary' type='submit'>
                        {this.state.mode === 'signin' ? 'Sign In' : 'Sign Up'}
                    </button>
                </div>
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

    private handleSubmit = event => {
        event.preventDefault()
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
                }, (err, authResult) => {
                    if (err) {
                        console.error('auth error: ', err)
                        this.setState({ errorDescription: err.description || 'Unknown Error' })
                    }
                })
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
                }, (err, authResult) => {
                    if (err) {
                        console.error('auth error: ', err)
                        this.setState({ errorDescription: err.description || 'Unknown Error' })
                    }
                })
                break
        }
    }
}

interface SignInPageProps {}

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class SignInPage extends React.Component<SignInPageProps> {

    public render(): JSX.Element | null {
        return (
            <div className='sign-in-page'>
                <PageTitle title='sign in or sign up' />
                <HeroPage icon={KeyIcon} title='Welcome to Sourcegraph' subtitle='Sign in or sign up to create an account' cta={<LoginSignupForm />} />
            </div>
        )
    }
}
