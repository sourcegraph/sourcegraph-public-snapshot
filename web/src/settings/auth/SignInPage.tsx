import KeyIcon from '@sourcegraph/icons/lib/Key'
import { WebAuth } from 'auth0-js'
import * as React from 'react'
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

interface Props {
    showEditorFlow: boolean
}

interface State {
    mode: 'signin' | 'signup'
    email: string
    password: string
    error: string
}

class LoginSignupForm extends React.Component<{}, State> {
    public state: State = {
        mode: 'signin',
        email: '',
        password: '',
        error: ''
    }

    public render(): JSX.Element | null {
        return (
            <form className='sign-in-page__form' onSubmit={this.handleSubmit}>
                {this.state.error !== '' &&
                    <p className='sign-in-page__error'>{this.state.error}</p>
                }
                <div>
                    <span className={`sign-in-page__mode${this.state.mode === 'signin' ? '--active' : ''}`} onClick={this.setModeSignIn}>Sign in</span>
                    <span className='sign-in-page__mode-divider'>|</span>
                    <span className={`sign-in-page__mode${this.state.mode === 'signup' ? '--active' : ''}`} onClick={this.setModeSignUp}>Sign up</span>
                </div>
                <input className='ui-text-input' onChange={this.onEmailFieldChange} value={this.state.email} type='email' placeholder='Email' />
                <input className='ui-text-input' onChange={this.onPasswordFieldChange} value={this.state.password} type='password' placeholder='Password' />
                <button className='ui-button' type='submit'>
                    {this.state.mode === 'signin' ? 'Sign In' : 'Sign Up'}
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
                        this.setState({ error: (err as any).description })
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
                        this.setState({ error: (err as any).description })
                    }
                })
                break
        }
    }
}

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class SignInPage extends React.Component<Props> {

    public render(): JSX.Element | null {
        return (
            <div className='sign-in-page'>
                <PageTitle title='sign in or sign up' />
                <HeroPage icon={KeyIcon} title='Welcome to Sourcegraph' subtitle='Sign in or sign up to create an account' cta={<LoginSignupForm />} />
            </div>
        )
    }
}
