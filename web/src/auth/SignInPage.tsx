import KeyIcon from '@sourcegraph/icons/lib/Key'
import * as H from 'history'
import * as React from 'react'
import { Redirect } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { getReturnTo } from './SignInSignUpCommon'
import { UsernamePasswordSignInForm } from './UsernamePasswordSignInForm'

interface SignInPageProps {
    location: H.Location
    history: H.History
    user: GQL.IUser | null
}

export class SignInPage extends React.Component<SignInPageProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('SignIn', {}, false)
    }

    public render(): JSX.Element | null {
        if (this.props.user) {
            const returnTo = getReturnTo(this.props.location)
            return <Redirect to={returnTo} />
        }

        return (
            <div className="signin-signup-page sign-in-page">
                <PageTitle title="Sign in" />
                <HeroPage
                    icon={KeyIcon}
                    title="Sign into Sourcegraph"
                    cta={<UsernamePasswordSignInForm {...this.props} />}
                />
            </div>
        )
    }
}
