import KeyIcon from '@sourcegraph/icons/lib/Key'
import * as React from 'react'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { SignInButton } from './SignInButton'

interface Props {
    showEditorFlow: boolean
}

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class SignInPage extends React.Component<Props> {

    public render(): JSX.Element | null {
        return (
            <div className='ui-section'>
                <PageTitle title='sign in or sign up' />
                <HeroPage icon={KeyIcon} title='Welcome to Sourcegraph' subtitle='Sign in or sign up to create an account' cta={<SignInButton />} />
            </div>
        )
    }
}
