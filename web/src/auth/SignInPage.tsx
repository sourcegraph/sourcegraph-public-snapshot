import * as H from 'history'
import KeyIcon from 'mdi-react/KeyIcon'
import * as React from 'react'
import { Redirect } from 'react-router-dom'
import * as GQL from '../../../shared/src/graphql/schema'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { getReturnTo } from './SignInSignUpCommon'
import { UsernamePasswordSignInForm } from './UsernamePasswordSignInForm'

interface SignInPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
}

export class SignInPage extends React.Component<SignInPageProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SignIn', false)
    }

    public render(): JSX.Element | null {
        if (this.props.authenticatedUser) {
            const returnTo = getReturnTo(this.props.location)
            return <Redirect to={returnTo} />
        }

        return (
            <div className="signin-signup-page sign-in-page">
                <PageTitle title="Sign in" />
                <HeroPage
                    icon={KeyIcon}
                    title="Sign into Sourcegraph"
                    body={
                        window.context.authProviders && window.context.authProviders.length > 0 ? (
                            <div className="mb-4">
                                {window.context.authProviders.map((provider, i) =>
                                    provider.isBuiltin ? (
                                        <UsernamePasswordSignInForm key={i} {...this.props} />
                                    ) : (
                                        <div className="mb-2">
                                            <a key={i} href={provider.authenticationURL} className="btn btn-secondary">
                                                Sign in with {provider.displayName}
                                            </a>
                                        </div>
                                    )
                                )}
                            </div>
                        ) : (
                            <div className="alert alert-info mt-3">
                                No authentication providers are available. Contact a site administrator for help.
                            </div>
                        )
                    }
                />
            </div>
        )
    }
}
