import * as H from 'history'
import * as React from 'react'
import { Link, Redirect } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { getReturnTo } from './SignInSignUpCommon'
import { SignUpArgs, SignUpForm } from './SignUpForm'
import { AuthenticatedUser } from '../auth'
import { SourcegraphIcon } from './icons'

interface SignUpPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
}

export class SignUpPage extends React.Component<SignUpPageProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SignUp', null, false)
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
                    icon={SourcegraphIcon}
                    iconLinkTo={window.context.sourcegraphDotComMode ? '/search' : undefined}
                    iconClassName="bg-transparent"
                    title={
                        window.context.sourcegraphDotComMode
                            ? 'Sign up for Sourcegraph Cloud'
                            : 'Sign up for Sourcegraph Server'
                    }
                    body={
                        <>
                            <p>
                                <Link to={`/sign-in${this.props.location.search}`}>
                                    Already have an account? Sign in.
                                </Link>
                            </p>
                            <SignUpForm {...this.props} doSignUp={this.doSignUp} />
                        </>
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
        }).then(response => {
            if (response.status !== 200) {
                return response.text().then(text => Promise.reject(new Error(text)))
            }
            window.location.replace(getReturnTo(this.props.location))
            return Promise.resolve()
        })
}
