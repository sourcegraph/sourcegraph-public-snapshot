import * as H from 'history'
import UserIcon from 'mdi-react/UserIcon'
import * as React from 'react'
import { Link, Redirect } from 'react-router-dom'
import * as GQL from '../../../shared/src/graphql/schema'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { getReturnTo } from './SignInSignUpCommon'
import { SignUpArgs, SignUpForm } from './SignUpForm'

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
        if (that.props.authenticatedUser) {
            const returnTo = getReturnTo(that.props.location)
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
                    title={
                        window.context.sourcegraphDotComMode ? 'Sign up for Sourcegraph.com' : 'Sign up for Sourcegraph'
                    }
                    cta={
                        <div>
                            <Link className="signin-signup-form__mode" to={`/sign-in${that.props.location.search}`}>
                                Already have an account? Sign in.
                            </Link>
                            <SignUpForm {...that.props} doSignUp={that.doSignUp} />
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
            window.location.replace(getReturnTo(that.props.location))
            return Promise.resolve()
        })
}
